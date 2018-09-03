// Code generated by counterfeiter. DO NOT EDIT.
package uaafakes

import (
	"sync"

	"github.com/cloudfoundry/bosh-cli/uaa"
)

type FakeUAA struct {
	NewStaleAccessTokenStub        func(refreshValue string) uaa.StaleAccessToken
	newStaleAccessTokenMutex       sync.RWMutex
	newStaleAccessTokenArgsForCall []struct {
		refreshValue string
	}
	newStaleAccessTokenReturns struct {
		result1 uaa.StaleAccessToken
	}
	newStaleAccessTokenReturnsOnCall map[int]struct {
		result1 uaa.StaleAccessToken
	}
	PromptsStub        func() ([]uaa.Prompt, error)
	promptsMutex       sync.RWMutex
	promptsArgsForCall []struct{}
	promptsReturns     struct {
		result1 []uaa.Prompt
		result2 error
	}
	promptsReturnsOnCall map[int]struct {
		result1 []uaa.Prompt
		result2 error
	}
	ClientCredentialsGrantStub        func() (uaa.Token, error)
	clientCredentialsGrantMutex       sync.RWMutex
	clientCredentialsGrantArgsForCall []struct{}
	clientCredentialsGrantReturns     struct {
		result1 uaa.Token
		result2 error
	}
	clientCredentialsGrantReturnsOnCall map[int]struct {
		result1 uaa.Token
		result2 error
	}
	OwnerPasswordCredentialsGrantStub        func([]uaa.PromptAnswer) (uaa.AccessToken, error)
	ownerPasswordCredentialsGrantMutex       sync.RWMutex
	ownerPasswordCredentialsGrantArgsForCall []struct {
		arg1 []uaa.PromptAnswer
	}
	ownerPasswordCredentialsGrantReturns struct {
		result1 uaa.AccessToken
		result2 error
	}
	ownerPasswordCredentialsGrantReturnsOnCall map[int]struct {
		result1 uaa.AccessToken
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeUAA) NewStaleAccessToken(refreshValue string) uaa.StaleAccessToken {
	fake.newStaleAccessTokenMutex.Lock()
	ret, specificReturn := fake.newStaleAccessTokenReturnsOnCall[len(fake.newStaleAccessTokenArgsForCall)]
	fake.newStaleAccessTokenArgsForCall = append(fake.newStaleAccessTokenArgsForCall, struct {
		refreshValue string
	}{refreshValue})
	fake.recordInvocation("NewStaleAccessToken", []interface{}{refreshValue})
	fake.newStaleAccessTokenMutex.Unlock()
	if fake.NewStaleAccessTokenStub != nil {
		return fake.NewStaleAccessTokenStub(refreshValue)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.newStaleAccessTokenReturns.result1
}

func (fake *FakeUAA) NewStaleAccessTokenCallCount() int {
	fake.newStaleAccessTokenMutex.RLock()
	defer fake.newStaleAccessTokenMutex.RUnlock()
	return len(fake.newStaleAccessTokenArgsForCall)
}

func (fake *FakeUAA) NewStaleAccessTokenArgsForCall(i int) string {
	fake.newStaleAccessTokenMutex.RLock()
	defer fake.newStaleAccessTokenMutex.RUnlock()
	return fake.newStaleAccessTokenArgsForCall[i].refreshValue
}

func (fake *FakeUAA) NewStaleAccessTokenReturns(result1 uaa.StaleAccessToken) {
	fake.NewStaleAccessTokenStub = nil
	fake.newStaleAccessTokenReturns = struct {
		result1 uaa.StaleAccessToken
	}{result1}
}

func (fake *FakeUAA) NewStaleAccessTokenReturnsOnCall(i int, result1 uaa.StaleAccessToken) {
	fake.NewStaleAccessTokenStub = nil
	if fake.newStaleAccessTokenReturnsOnCall == nil {
		fake.newStaleAccessTokenReturnsOnCall = make(map[int]struct {
			result1 uaa.StaleAccessToken
		})
	}
	fake.newStaleAccessTokenReturnsOnCall[i] = struct {
		result1 uaa.StaleAccessToken
	}{result1}
}

func (fake *FakeUAA) Prompts() ([]uaa.Prompt, error) {
	fake.promptsMutex.Lock()
	ret, specificReturn := fake.promptsReturnsOnCall[len(fake.promptsArgsForCall)]
	fake.promptsArgsForCall = append(fake.promptsArgsForCall, struct{}{})
	fake.recordInvocation("Prompts", []interface{}{})
	fake.promptsMutex.Unlock()
	if fake.PromptsStub != nil {
		return fake.PromptsStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.promptsReturns.result1, fake.promptsReturns.result2
}

func (fake *FakeUAA) PromptsCallCount() int {
	fake.promptsMutex.RLock()
	defer fake.promptsMutex.RUnlock()
	return len(fake.promptsArgsForCall)
}

func (fake *FakeUAA) PromptsReturns(result1 []uaa.Prompt, result2 error) {
	fake.PromptsStub = nil
	fake.promptsReturns = struct {
		result1 []uaa.Prompt
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) PromptsReturnsOnCall(i int, result1 []uaa.Prompt, result2 error) {
	fake.PromptsStub = nil
	if fake.promptsReturnsOnCall == nil {
		fake.promptsReturnsOnCall = make(map[int]struct {
			result1 []uaa.Prompt
			result2 error
		})
	}
	fake.promptsReturnsOnCall[i] = struct {
		result1 []uaa.Prompt
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) ClientCredentialsGrant() (uaa.Token, error) {
	fake.clientCredentialsGrantMutex.Lock()
	ret, specificReturn := fake.clientCredentialsGrantReturnsOnCall[len(fake.clientCredentialsGrantArgsForCall)]
	fake.clientCredentialsGrantArgsForCall = append(fake.clientCredentialsGrantArgsForCall, struct{}{})
	fake.recordInvocation("ClientCredentialsGrant", []interface{}{})
	fake.clientCredentialsGrantMutex.Unlock()
	if fake.ClientCredentialsGrantStub != nil {
		return fake.ClientCredentialsGrantStub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.clientCredentialsGrantReturns.result1, fake.clientCredentialsGrantReturns.result2
}

func (fake *FakeUAA) ClientCredentialsGrantCallCount() int {
	fake.clientCredentialsGrantMutex.RLock()
	defer fake.clientCredentialsGrantMutex.RUnlock()
	return len(fake.clientCredentialsGrantArgsForCall)
}

func (fake *FakeUAA) ClientCredentialsGrantReturns(result1 uaa.Token, result2 error) {
	fake.ClientCredentialsGrantStub = nil
	fake.clientCredentialsGrantReturns = struct {
		result1 uaa.Token
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) ClientCredentialsGrantReturnsOnCall(i int, result1 uaa.Token, result2 error) {
	fake.ClientCredentialsGrantStub = nil
	if fake.clientCredentialsGrantReturnsOnCall == nil {
		fake.clientCredentialsGrantReturnsOnCall = make(map[int]struct {
			result1 uaa.Token
			result2 error
		})
	}
	fake.clientCredentialsGrantReturnsOnCall[i] = struct {
		result1 uaa.Token
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrant(arg1 []uaa.PromptAnswer) (uaa.AccessToken, error) {
	var arg1Copy []uaa.PromptAnswer
	if arg1 != nil {
		arg1Copy = make([]uaa.PromptAnswer, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.ownerPasswordCredentialsGrantMutex.Lock()
	ret, specificReturn := fake.ownerPasswordCredentialsGrantReturnsOnCall[len(fake.ownerPasswordCredentialsGrantArgsForCall)]
	fake.ownerPasswordCredentialsGrantArgsForCall = append(fake.ownerPasswordCredentialsGrantArgsForCall, struct {
		arg1 []uaa.PromptAnswer
	}{arg1Copy})
	fake.recordInvocation("OwnerPasswordCredentialsGrant", []interface{}{arg1Copy})
	fake.ownerPasswordCredentialsGrantMutex.Unlock()
	if fake.OwnerPasswordCredentialsGrantStub != nil {
		return fake.OwnerPasswordCredentialsGrantStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fake.ownerPasswordCredentialsGrantReturns.result1, fake.ownerPasswordCredentialsGrantReturns.result2
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantCallCount() int {
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	return len(fake.ownerPasswordCredentialsGrantArgsForCall)
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantArgsForCall(i int) []uaa.PromptAnswer {
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	return fake.ownerPasswordCredentialsGrantArgsForCall[i].arg1
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantReturns(result1 uaa.AccessToken, result2 error) {
	fake.OwnerPasswordCredentialsGrantStub = nil
	fake.ownerPasswordCredentialsGrantReturns = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) OwnerPasswordCredentialsGrantReturnsOnCall(i int, result1 uaa.AccessToken, result2 error) {
	fake.OwnerPasswordCredentialsGrantStub = nil
	if fake.ownerPasswordCredentialsGrantReturnsOnCall == nil {
		fake.ownerPasswordCredentialsGrantReturnsOnCall = make(map[int]struct {
			result1 uaa.AccessToken
			result2 error
		})
	}
	fake.ownerPasswordCredentialsGrantReturnsOnCall[i] = struct {
		result1 uaa.AccessToken
		result2 error
	}{result1, result2}
}

func (fake *FakeUAA) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.newStaleAccessTokenMutex.RLock()
	defer fake.newStaleAccessTokenMutex.RUnlock()
	fake.promptsMutex.RLock()
	defer fake.promptsMutex.RUnlock()
	fake.clientCredentialsGrantMutex.RLock()
	defer fake.clientCredentialsGrantMutex.RUnlock()
	fake.ownerPasswordCredentialsGrantMutex.RLock()
	defer fake.ownerPasswordCredentialsGrantMutex.RUnlock()
	return fake.invocations
}

func (fake *FakeUAA) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ uaa.UAA = new(FakeUAA)
