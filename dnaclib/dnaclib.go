package dnaclib

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	dnac "github.com/cisco-en-programmability/dnacenter-go-sdk/v4/sdk"
	"golang.org/x/term"
)

func promptForLogin() (string, string, string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter DNAC IP Address: ")
	dnacIP, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}

	fmt.Print("Enter Username: ")
	username, err := reader.ReadString('\n')
	if err != nil {
		return "", "", "", err
	}

	fmt.Print("Enter Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", "", "", err
	}

	password := string(bytePassword)
	return strings.TrimSpace(dnacIP), strings.TrimSpace(username), strings.TrimSpace(password), nil
}

func promptForString(prompt string) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt, ": ")
	n, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(n), nil
}

func getDeviceByIP(Client *dnac.Client, ipAddress string) (string, error) {
	device, _, err := Client.Devices.GetNetworkDeviceByIP(ipAddress)
	if err != nil {
		return "", err
	}
	return device.Response.ID, nil
}

func LoginToDNAC() (*dnac.Client, error) {
	dnacIP, username, password, err := promptForLogin()
	if err != nil {
		return nil, err
	}
	url := "https://{ip}"
	return dnac.NewClientWithOptions(
		strings.Replace(url, "{ip}", dnacIP, -1),
		username, password,
		"false",
		"false")

}

func generateSingleTargetRenameInfo(deviceID string, newName string) *dnac.RequestConfigurationTemplatesDeployTemplateTargetInfo {
	return &dnac.RequestConfigurationTemplatesDeployTemplateTargetInfo{
		ID: deviceID,
		Params: &dnac.RequestConfigurationTemplatesDeployTemplateTargetInfoParams{
			"hostname": newName,
		},
		Type: "MANAGED_DEVICE_UUID",
	}
}

func generateTemplateDeployment(templateID string, targets *[]dnac.RequestConfigurationTemplatesDeployTemplateTargetInfo) *dnac.RequestConfigurationTemplatesDeployTemplate {
	forceTemplate := false
	return &dnac.RequestConfigurationTemplatesDeployTemplate{
		ForcePushTemplate: &forceTemplate,
		TemplateID:        templateID,
		TargetInfo:        targets,
	}
}

func getTemplateIDByName(Client *dnac.Client, name string) (string, error) {
	templates, _, err := Client.ConfigurationTemplates.GetTemplatesDetails(
		&dnac.GetTemplatesDetailsQueryParams{
			Name: name,
		},
	)
	if err != nil {
		return "", err
	}
	var templateID string
	templateID = ""
	for _, template := range *templates.Response {
		templateID = template.ID

	}
	return templateID, nil
}

func waitForDeployment(Client *dnac.Client, deploymentID string) {
	for i, _, err := Client.ConfigurationTemplates.StatusOfTemplateDeployment(deploymentID); i.Status != "SUCCESS"; i, _, err = Client.ConfigurationTemplates.StatusOfTemplateDeployment(deploymentID) {
		if err != nil {
			panic(err)
		}
		if i.Status == "FAILURE" {
			panic(i.Status)
		}
		fmt.Printf("%v %v %v\r", i.Status, i.StatusMessage, i.Duration)
	}
}

func RenameDevice(Client *dnac.Client) (string, error) {
	deviceIP, err := promptForString("\nPlease enter DeviceIP")
	if err != nil {
		return "", err
	}
	deviceID, err := getDeviceByIP(Client, deviceIP)
	if err != nil {
		return "", err
	}
	newName, err := promptForString("Please enter New Hostname")
	if err != nil {
		return "", err
	}
	targets := []dnac.RequestConfigurationTemplatesDeployTemplateTargetInfo{*generateSingleTargetRenameInfo(deviceID, newName)}
	templateID, err := getTemplateIDByName(Client, "rename")
	if err != nil {
		fmt.Println(err)
	}
	deploymentID, _, err := Client.ConfigurationTemplates.DeployTemplate(generateTemplateDeployment(templateID, &targets))
	if err != nil {
		panic(err)
	}
	parsedDeploymentID := strings.TrimSpace(strings.Split(deploymentID.DeploymentID, ":")[3])
	waitForDeployment(Client, parsedDeploymentID)
	return "SUCCESS", nil
}
