package cmd

import (
	"net/rpc"
	"time"

	"github.com/vscode-lcode/hub/cmd/hub/fl"
	"golang.org/x/net/webdav"
)

type RemoteLockSystem struct {
	client *rpc.Client
}

func NewRemoteLockSystem(client *rpc.Client) *RemoteLockSystem {
	return &RemoteLockSystem{client: client}
}

var _ webdav.LockSystem = (*RemoteLockSystem)(nil)

func (ls *RemoteLockSystem) Confirm(now time.Time, name0, name1 string, conditions ...webdav.Condition) (release func(), err error) {
	params := fl.ConfirmParams{Now: now, Name0: name0, Name1: name1, Conditions: conditions}
	var callback uintptr
	err = ls.client.Call("LockSystem.Confirm", params, &callback)
	if err != nil {
		return nil, err
	}
	release = func() {
		ls.client.Call("LockSystem.ConfirmCallback", callback, nil)
	}
	return release, nil
}
func (ls *RemoteLockSystem) Create(now time.Time, details webdav.LockDetails) (token string, err error) {
	params := fl.CreateParams{Now: now, Details: details}
	err = ls.client.Call("LockSystem.Create", params, &token)
	if err != nil {
		return "", err
	}
	return token, nil
}
func (ls *RemoteLockSystem) Refresh(now time.Time, token string, duration time.Duration) (ld webdav.LockDetails, err error) {
	params := fl.RefreshParams{Now: now, Token: token, Duration: duration}
	err = ls.client.Call("LockSystem.Refresh", params, &ld)
	if err != nil {
		return ld, err
	}
	return ld, err
}

func (ls *RemoteLockSystem) Unlock(now time.Time, token string) (err error) {
	params := fl.UnlockParams{Now: now, Token: token}
	var reply bool
	err = ls.client.Call("LockSystem.Unlock", params, &reply)
	if err != nil {
		return err
	}
	return err
}
