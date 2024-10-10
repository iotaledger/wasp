package move

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func BuildPackage(contractPath string) (*PackageBytecode, error) {
	var err error
	cmd := exec.Command("iotago", "move", "build", "--dump-bytecode-as-base64")
	// TODO skip to fetch latest deps if there is no internet
	// cmd := exec.Command("iotago", "move", "build", "--dump-bytecode-as-base64", "--skip-fetch-latest-git-deps")
	cmd.Dir = contractPath

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to execute iotago cli: %w, stderr: '%s'", err, stderr.String())
	}

	var modules PackageBytecode
	err = json.Unmarshal(stdout.Bytes(), &modules)
	if err != nil {
		return nil, fmt.Errorf("failed to parse move build result: %w", err)
	}

	return &modules, nil
}
