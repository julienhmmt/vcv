package vault_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"vcv/internal/vault"
)

func TestCheckInstances_ParallelAndNil(t *testing.T) {
	fast := &vault.MockClient{}
	fast.On("CheckConnection", mock.Anything).Return(nil).Once()
	slow := &vault.MockClient{}
	slow.On("CheckConnection", mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(50 * time.Millisecond)
	}).Return(nil).Once()
	failing := &vault.MockClient{}
	failing.On("CheckConnection", mock.Anything).Return(errors.New("down")).Once()

	clients := map[string]vault.Client{
		"a": fast,
		"b": slow,
		"c": failing,
	}
	start := time.Now()
	results := vault.CheckInstances(context.Background(), []string{"a", "b", "c", "missing"}, clients, 2*time.Second)
	elapsed := time.Since(start)
	require.Len(t, results, 4)
	assert.True(t, results[0].Connected)
	assert.True(t, results[1].Connected)
	assert.False(t, results[2].Connected)
	assert.Error(t, results[2].Error)
	assert.False(t, results[3].Connected)
	assert.Less(t, elapsed, 500*time.Millisecond, "checks should run in parallel")
	fast.AssertExpectations(t)
	slow.AssertExpectations(t)
	failing.AssertExpectations(t)
}

func TestCheckInstances_Timeout(t *testing.T) {
	hung := &vault.MockClient{}
	hung.On("CheckConnection", mock.Anything).Run(func(args mock.Arguments) {
		ctx := args.Get(0).(context.Context)
		<-ctx.Done()
	}).Return(context.DeadlineExceeded).Once()

	results := vault.CheckInstances(context.Background(), []string{"hung"}, map[string]vault.Client{"hung": hung}, 30*time.Millisecond)
	require.Len(t, results, 1)
	assert.False(t, results[0].Connected)
	assert.Error(t, results[0].Error)
	hung.AssertExpectations(t)
}
