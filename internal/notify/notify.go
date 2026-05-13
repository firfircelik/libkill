package notify

import (
	"os/exec"
)

type Notifier struct{}

func New() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(title, message string) error {
	return sendNotification(title, message)
}

func isCommandAvailable(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
