package git

import (
	"fmt"
	"regexp"
	"slices"
)

type Identity struct {
	Name    string   `json:"-"` // Identity name (map key)
	User    string   `json:"user"`
	Email   string   `json:"email"`
	Domain  string   `json:"domain"`
	Folders []string `json:"folders"`
}

func (i *Identity) Validate() error {
	if i.User == "" {
		return fmt.Errorf("user must not be empty")
	}
	if i.Email == "" {
		return fmt.Errorf("email must not be empty")
	}
	if i.Domain == "" {
		return fmt.Errorf("domain must not be empty")
	}
	if len(i.Folders) == 0 {
		return fmt.Errorf("at least one folder must be specified")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(i.Email) {
		return fmt.Errorf("invalid email address: %s", i.Email)
	}

	if slices.Contains(i.Folders, "") {
		return fmt.Errorf("folder path must not be empty")
	}

	return nil
}

func (i *Identity) SSHKeyPath() string {
	return fmt.Sprintf("~/.ssh/%s_key", i.Name)
}

func (i *Identity) SSHPubKeyPath() string {
	return fmt.Sprintf("~/.ssh/%s_key.pub", i.Name)
}

func (i *Identity) GitConfigPath() string {
	return fmt.Sprintf("~/.gitconfig-%s", i.Name)
}

func (i *Identity) SSHKeyComment() string {
	return fmt.Sprintf("%s [zzk:%s]", i.Email, i.Name)
}
