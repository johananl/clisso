package okta

import (
	"fmt"
	"log"
	"time"

	"github.com/allcloud-io/clisso/aws"
	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/provider"
	"github.com/allcloud-io/clisso/saml"
	"github.com/allcloud-io/clisso/spinner"
	"github.com/fatih/color"
)

const (
	MFATypePush = "push"
	MFATypeTOTP = "token:software:totp"

	VerifyFactorStatusSuccess = "SUCCESS"
	VerifyFactorStatusWaiting = "WAITING"
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
func (p *Provider) Get(app provider.App, duration int64) (*aws.Credentials, error) {
	// Get user credentials
	user := p.Config.Username
	if user == "" {
		// Get credentials from the user
		fmt.Print("Okta username: ")
		fmt.Scanln(&user)
	}

	pass, err := keyChain.Get(p.Name)

	// Initialize spinner
	var s = spinner.New()

	// Get session token
	s.Start()
	resp, err := p.Client.GetSessionToken(&GetSessionTokenParams{
		Username: user,
		Password: string(pass),
	})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("getting session token: %v", err)
	}

	var st string

	// TODO Handle multiple MFA devices (allow user to choose)
	switch resp.Status {
	case StatusSuccess:
		st = resp.SessionToken
	case StatusMFARequired:
		factor := resp.Embedded.Factors[0]
		stateToken := resp.StateToken

		var vfResp *VerifyFactorResponse

		switch factor.FactorType {
		case MFATypePush:
			// Okta Verify push notification:
			// https://developer.okta.com/docs/api/resources/authn/#verify-push-factor
			// Keep polling authentication transactions with WAITING result until the challenge
			// completes or expires.
			fmt.Println("Please approve request on Okta Verify app")
			s.Start()
			vfResp, err = p.Client.VerifyFactor(&VerifyFactorParams{
				FactorID:   factor.ID,
				StateToken: stateToken,
			})
			if err != nil {
				return nil, fmt.Errorf("verifying MFA: %v", err)
			}

			for vfResp.FactorResult == VerifyFactorStatusWaiting {
				vfResp, err = p.Client.VerifyFactor(&VerifyFactorParams{
					FactorID:   factor.ID,
					StateToken: stateToken,
				})
				time.Sleep(2 * time.Second)
			}
			s.Stop()
		case MFATypeTOTP:
			fmt.Print("Please enter the OTP from your MFA device: ")
			var otp string
			fmt.Scanln(&otp)

			s.Start()
			vfResp, err = p.Client.VerifyFactor(&VerifyFactorParams{
				FactorID:   factor.ID,
				PassCode:   otp,
				StateToken: stateToken,
			})
			s.Stop()
		default:
			return nil, fmt.Errorf("unsupported MFA type '%s'", factor.FactorType)
		}

		if err != nil {
			return nil, fmt.Errorf("verifying MFA: %v", err)
		}

		// Handle failed MFA verification (verification rejected or timed out)
		if vfResp.Status != VerifyFactorStatusSuccess {
			return nil, fmt.Errorf("MFA verification failed")
		}

		st = vfResp.SessionToken
	default:
		return nil, fmt.Errorf("Invalid status %s", resp.Status)
	}

	// Launch Okta app with session token
	s.Start()
	samlAssertion, err := p.Client.LaunchApp(&LaunchAppParams{SessionToken: st, URL: app.ID()})
	s.Stop()
	if err != nil {
		return nil, fmt.Errorf("Error launching app: %v", err)
	}

	arn, err := saml.Get(*samlAssertion)
	if err != nil {
		return nil, err
	}

	s.Start()
	creds, err := aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, duration)
	s.Stop()

	if err != nil {
		if err.Error() == aws.ErrDurationExceeded {
			log.Println(color.YellowString(aws.DurationExceededMessage))
			s.Start()
			creds, err = aws.AssumeSAMLRole(arn.Provider, arn.Role, *samlAssertion, 3600)
			s.Stop()
		}
	}

	return creds, err
}
