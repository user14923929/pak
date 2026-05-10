package dpkg

import (
	"fmt"
	"os/exec"
)

// Install calls dpkg -i on the given .deb file.
func Install(debPath string) error {
	cmd := exec.Command("dpkg", "-i", debPath)
	cmd.Stdout = nil // suppress dpkg chatter; caller may redirect
	cmd.Stderr = nil
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dpkg -i %s: %w\n%s", debPath, err, out)
	}
	return nil
}

// Remove calls dpkg --remove (keeps config files) on the given package name.
func Remove(name string) error {
	cmd := exec.Command("dpkg", "--remove", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dpkg --remove %s: %w\n%s", name, err, out)
	}
	return nil
}

// Purge calls dpkg --purge (removes config files too).
func Purge(name string) error {
	cmd := exec.Command("dpkg", "--purge", name)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("dpkg --purge %s: %w\n%s", name, err, out)
	}
	return nil
}
