package onelogin

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/allcloud-io/clisso/keychain"
	"github.com/allcloud-io/clisso/platform/aws"
	"github.com/allcloud-io/clisso/provider"
	"github.com/allcloud-io/clisso/saml"
	"github.com/fatih/color"
)

const (
	// MFADeviceOneLoginProtect symbolizes the OneLogin Protect mobile app, which supports push
	// notifications. More info here: https://developers.onelogin.com/api-docs/1/saml-assertions/verify-factor
	MFADeviceOneLoginProtect = "OneLogin Protect"

	// MFAPushTimeout represents the number of seconds to wait for a successful push attempt before
	// falling back to OTP input.
	MFAPushTimeout = 30

	// MFAInterval represents the interval at which we check for an accepted push message.
	MFAInterval = 1
)

var (
	keyChain = keychain.DefaultKeychain{}
)

// Get gets temporary credentials for the given app.
// TODO Move AWS logic outside this function.
func (p *Provider) Get(user string, pass string, app provider.App, duration int64) (*aws.Credentials, error) {
	// Get OneLogin access token
	token, err := p.Client.GenerateTokens(p.Config.ClientID, p.Config.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("generating access token: %s", err)
	}

	// Generate SAML assertion
	pSAML := GenerateSamlAssertionParams{
		UsernameOrEmail: user,
		Password:        string(pass),
		AppId:           app.ID(),
		// TODO At the moment when there is a mismatch between Subdomain and
		// the domain in the username, the user is getting HTTP 400.
		Subdomain: p.Config.Subdomain,
	}

	rSaml, err := p.Client.GenerateSamlAssertion(token, &pSAML)
	if err != nil {
		return nil, fmt.Errorf("generating SAML assertion: %v", err)
	}

	st := rSaml.Data[0].StateToken

	devices := rSaml.Data[0].Devices
	device, err := getDevice(devices)

	var rMfa *VerifyFactorResponse

	var pushOK = false

	if device.DeviceType == MFADeviceOneLoginProtect {
		// Push is supported by the selected MFA device - try pushing and fall back to manual input
		pushOK = true
		pMfa := VerifyFactorParams{
			AppId:       app.ID(),
			DeviceId:    fmt.Sprintf("%v", device.DeviceID),
			StateToken:  st,
			OtpToken:    "",
			DoNotNotify: false,
		}

		rMfa, err = p.Client.VerifyFactor(token, &pMfa)
		if err != nil {
			return nil, err
		}

		pMfa.DoNotNotify = true

		fmt.Println(rMfa.Status.Message)

		timeout := MFAPushTimeout
		for rMfa.Status.Type == "pending" && timeout > 0 {
			time.Sleep(time.Duration(MFAInterval) * time.Second)
			rMfa, err = p.Client.VerifyFactor(token, &pMfa)
			if err != nil {
				return nil, err
			}

			timeout -= MFAInterval
		}

		if rMfa.Status.Type == "pending" {
			fmt.Println("MFA verification timed out - falling back to manual OTP input")
			pushOK = false
		}
	}

	if !pushOK {
		// Push failed or not supported by the selected MFA device
		fmt.Print("Please enter the OTP from your MFA device: ")
		var otp string
		fmt.Scanln(&otp)

		// Verify MFA
		pMfa := VerifyFactorParams{
			AppId:       app.ID(),
			DeviceId:    fmt.Sprintf("%v", device.DeviceID),
			StateToken:  st,
			OtpToken:    otp,
			DoNotNotify: false,
		}

		rMfa, err = p.Client.VerifyFactor(token, &pMfa)
		if err != nil {
			return nil, fmt.Errorf("verifying factor: %v", err)
		}
	}

	arn, err := saml.Get(rMfa.Data)
	if err != nil {
		return nil, err
	}

	creds, err := aws.AssumeSAMLRole(arn.Provider, arn.Role, rMfa.Data, duration)
	if err != nil {
		if err.Error() == aws.ErrDurationExceeded {
			log.Println(color.YellowString(aws.DurationExceededMessage))
			creds, err = aws.AssumeSAMLRole(arn.Provider, arn.Role, rMfa.Data, 3600)
		}
	}

	return creds, err
}

// getDevice gets a slice of MFA devices, prompts the user to select one and returns the selected device.
// If the slice contains only a single device, that device is returned. If the slice is empty, an error is returned.
// TODO: Move interactive prompts out of this function.
func getDevice(devices []Device) (device *Device, err error) {
	if len(devices) == 0 {
		// This should never happen
		err = errors.New("No MFA device returned by Onelogin")
		return
	}

	if len(devices) == 1 {
		device = &Device{DeviceID: devices[0].DeviceID, DeviceType: devices[0].DeviceType}
		return
	}

	var selection int
	for {
		for i, d := range devices {
			fmt.Printf("%d. %d - %s\n", i+1, d.DeviceID, d.DeviceType)
		}

		fmt.Printf("Please choose an MFA device to authenticate with (1-%d): ", len(devices))
		var input string
		_, err := fmt.Scanln(&input)
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			continue
		}

		// Verify we got an integer.
		selection, err = strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Invalid input '%s'\n", input)
			continue
		}

		// Verify selection is within range.
		if selection < 1 || selection > len(devices) {
			fmt.Printf("Invalid value %d. Valid values: 1-%d\n", selection, len(devices))
			continue
		}
		break
	}
	device = &Device{DeviceID: devices[selection-1].DeviceID, DeviceType: devices[selection-1].DeviceType}
	return
}
