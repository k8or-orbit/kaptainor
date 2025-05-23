package exec

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/dataservices"
	"github.com/portainer/portainer/api/stacks/stackutils"

	"github.com/rs/zerolog/log"
	"github.com/segmentio/encoding/json"
)

// SwarmStackManager represents a service for managing stacks.
type SwarmStackManager struct {
	binaryPath           string
	configPath           string
	signatureService     portainer.DigitalSignatureService
	fileService          portainer.FileService
	reverseTunnelService portainer.ReverseTunnelService
	dataStore            dataservices.DataStore
}

// NewSwarmStackManager initializes a new SwarmStackManager service.
// It also updates the configuration of the Docker CLI binary.
func NewSwarmStackManager(
	binaryPath, configPath string,
	signatureService portainer.DigitalSignatureService,
	fileService portainer.FileService,
	reverseTunnelService portainer.ReverseTunnelService,
	datastore dataservices.DataStore,
) (*SwarmStackManager, error) {
	manager := &SwarmStackManager{
		binaryPath:           binaryPath,
		configPath:           configPath,
		signatureService:     signatureService,
		fileService:          fileService,
		reverseTunnelService: reverseTunnelService,
		dataStore:            datastore,
	}

	if err := manager.updateDockerCLIConfiguration(manager.configPath); err != nil {
		return nil, err
	}

	return manager, nil
}

// Login executes the docker login command against a list of registries (including DockerHub).
func (manager *SwarmStackManager) Login(registries []portainer.Registry, endpoint *portainer.Endpoint) error {
	command, args, err := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.configPath, endpoint)
	if err != nil {
		return err
	}

	for _, registry := range registries {
		if registry.Authentication {
			username, password, err := getEffectiveRegUsernamePassword(manager.dataStore, &registry)
			if err != nil {
				continue
			}

			registryArgs := append(args, "login", "--username", username, "--password", password, registry.URL)
			if err := runCommandAndCaptureStdErr(command, registryArgs, nil, ""); err != nil {
				log.Warn().
					Err(err).
					Str("RegistryName", registry.Name).
					Msg("Failed to login.")
			}
		}
	}

	return nil
}

// Logout executes the docker logout command.
func (manager *SwarmStackManager) Logout(endpoint *portainer.Endpoint) error {
	command, args, err := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.configPath, endpoint)
	if err != nil {
		return err
	}

	args = append(args, "logout")

	return runCommandAndCaptureStdErr(command, args, nil, "")
}

// Deploy executes the docker stack deploy command.
func (manager *SwarmStackManager) Deploy(stack *portainer.Stack, prune bool, pullImage bool, endpoint *portainer.Endpoint) error {
	filePaths := stackutils.GetStackFilePaths(stack, true)
	command, args, err := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.configPath, endpoint)
	if err != nil {
		return err
	}

	if prune {
		args = append(args, "stack", "deploy", "--prune", "--with-registry-auth")
	} else {
		args = append(args, "stack", "deploy", "--with-registry-auth")
	}

	if !pullImage {
		args = append(args, "--resolve-image=never")
	}

	args = configureFilePaths(args, filePaths)
	args = append(args, stack.Name)

	env := make([]string, 0)
	for _, envvar := range stack.Env {
		env = append(env, envvar.Name+"="+envvar.Value)
	}

	return runCommandAndCaptureStdErr(command, args, env, stack.ProjectPath)
}

// Remove executes the docker stack rm command.
func (manager *SwarmStackManager) Remove(stack *portainer.Stack, endpoint *portainer.Endpoint) error {
	command, args, err := manager.prepareDockerCommandAndArgs(manager.binaryPath, manager.configPath, endpoint)
	if err != nil {
		return err
	}

	args = append(args, "stack", "rm", "--detach=false", stack.Name)

	return runCommandAndCaptureStdErr(command, args, nil, "")
}

func runCommandAndCaptureStdErr(command string, args []string, env []string, workingDir string) error {
	var stderr bytes.Buffer

	cmd := exec.Command(command, args...)
	cmd.Stderr = &stderr

	if workingDir != "" {
		cmd.Dir = workingDir
	}

	if env != nil {
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, env...)
	}

	if err := cmd.Run(); err != nil {
		return errors.New(stderr.String())
	}

	return nil
}

func (manager *SwarmStackManager) prepareDockerCommandAndArgs(binaryPath, configPath string, endpoint *portainer.Endpoint) (string, []string, error) {
	// Assume Linux as a default
	command := path.Join(binaryPath, "docker")

	if runtime.GOOS == "windows" {
		command = path.Join(binaryPath, "docker.exe")
	}

	args := make([]string, 0)
	args = append(args, "--config", configPath)

	endpointURL := endpoint.URL
	if endpoint.Type == portainer.EdgeAgentOnDockerEnvironment {
		tunnelAddr, err := manager.reverseTunnelService.TunnelAddr(endpoint)
		if err != nil {
			return "", nil, err
		}

		endpointURL = "tcp://" + tunnelAddr
	}

	args = append(args, "-H", endpointURL)

	if endpoint.TLSConfig.TLS {
		args = append(args, "--tls")

		if !endpoint.TLSConfig.TLSSkipVerify {
			args = append(args, "--tlsverify", "--tlscacert", endpoint.TLSConfig.TLSCACertPath)
		} else {
			args = append(args, "--tlscacert", "''")
		}

		if endpoint.TLSConfig.TLSCertPath != "" && endpoint.TLSConfig.TLSKeyPath != "" {
			args = append(args, "--tlscert", endpoint.TLSConfig.TLSCertPath, "--tlskey", endpoint.TLSConfig.TLSKeyPath)
		}
	}

	return command, args, nil
}

func (manager *SwarmStackManager) updateDockerCLIConfiguration(configPath string) error {
	configFilePath := path.Join(configPath, "config.json")

	config, err := manager.retrieveConfigurationFromDisk(configFilePath)
	if err != nil {
		log.Warn().Err(err).Msg("unable to retrieve the Swarm configuration from disk, proceeding without it")
	}

	signature, err := manager.signatureService.CreateSignature(portainer.PortainerAgentSignatureMessage)
	if err != nil {
		return err
	}

	if config["HttpHeaders"] == nil {
		config["HttpHeaders"] = make(map[string]any)
	}

	headersObject := config["HttpHeaders"].(map[string]any)
	headersObject["X-PortainerAgent-ManagerOperation"] = "1"
	headersObject["X-PortainerAgent-Signature"] = signature
	headersObject["X-PortainerAgent-PublicKey"] = manager.signatureService.EncodedPublicKey()

	return manager.fileService.WriteJSONToFile(configFilePath, config)
}

func (manager *SwarmStackManager) retrieveConfigurationFromDisk(path string) (map[string]any, error) {
	var config map[string]any

	raw, err := manager.fileService.GetFileContent(path, "")
	if err != nil {
		return make(map[string]any), nil
	}

	if err := json.Unmarshal(raw, &config); err != nil {
		return nil, err
	}

	return config, nil
}

func (manager *SwarmStackManager) NormalizeStackName(name string) string {
	return stackNameNormalizeRegex.ReplaceAllString(strings.ToLower(name), "")
}

func configureFilePaths(args []string, filePaths []string) []string {
	for _, path := range filePaths {
		args = append(args, "--compose-file", path)
	}

	return args
}
