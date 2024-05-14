package cedar

import (
	_ "embed"
	"encoding/json"
	"log"
	"sync"
	"testing"
)

type PolicyDataBucket struct {
	Policies             string  `json:"policies"`
	Entities             string  `json:"entities"`
	Request              Request `json:"request"`
	PolicyText           string  `json:"policyText"`
	VO                   string  `json:"vo"`
	Id                   string  `json:"id"`
	ErrorExpected        bool    `json:"errorExpected"`
	ExpectedErrorMessage string  `json:"expectedErrorMessage"`
}

var (
	once     sync.Once
	SuiteMap = make(map[string]PolicyDataBucket)
)

//go:embed josnparser_test_data.json
var SuiteDataFile string

func setup() {
	var d []PolicyDataBucket
	err := json.Unmarshal([]byte(SuiteDataFile), &d)
	if err != nil {
		log.Fatalf("Error unmarshalling Test Data JSON file: %v", err)
	}
	for _, item := range d {
		SuiteMap[item.Id] = item
	}
}

func execTest(t *testing.T, testId string, request Request) {
	once.Do(setup)
	testSubject, ok := SuiteMap[testId]
	if !ok {
		t.Logf("failed to fetch test data pertaining to testId: %s", testId)
		t.FailNow()
	}
	testSubject.Request = request
	text, err := new(PolicyParser).JsonToPolicy([]byte(testSubject.Policies))
	if err != nil && !testSubject.ErrorExpected {
		t.Log(text)
		t.Errorf("Test %s failed: Unexpected error while constructing Policy Text from JSON %v", testId, err)
		return
	} else if err != nil && testSubject.ErrorExpected {
		return
	}
	testSubject.PolicyText = text
	ps, err := NewPolicySet(testId, []byte(text))
	if err != nil && !testSubject.ErrorExpected {
		t.Log(text)
		t.Errorf("Test %s failed: Unexpected error while constructing NewPolicySet: %v", testId, err)
		return
	} else if err == nil && testSubject.ErrorExpected {
		t.Log(text)
		t.Errorf("Test %s failed: Expected error while constructing NewPolicySet, but got none", testId)
		return
	}
	if testSubject.ErrorExpected {
		return
	}
	var entities Entities
	if err := json.Unmarshal([]byte(testSubject.Entities), &entities); err != nil {
		t.Fatal(err)
	}
	d, _ := ps.IsAuthorized(entities, testSubject.Request)
	if d.String() != testSubject.VO {
		t.Errorf("expected %q but got %q", testSubject.VO, d.String())
	}
}

func TestBasicParsing(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "007"},
		Action:    EntityUID{Type: "Action", ID: "clean"},
		Resource:  EntityUID{Type: "Kitchen", ID: "kitchen"},
		Context:   nil,
	}
	testId := "simple-policy"
	execTest(t, testId, request)
}

func TestAdminDeleteFile(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "admin"},
		Action:    EntityUID{Type: "Action", ID: "delete"},
		Resource:  EntityUID{Type: "File", ID: "confidential"},
		Context:   nil,
	}
	testId := "admin-delete-file"
	execTest(t, testId, request)
}

func TestGuestWriteReadonly(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "002"},
		Action:    EntityUID{Type: "Action", ID: "write"},
		Resource:  EntityUID{Type: "File", ID: "readonly"},
		Context:   nil,
	}
	testId := "guest-write-readonly"
	execTest(t, testId, request)
}

func TestConditionalAccess(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "user123"},
		Action:    EntityUID{Type: "Action", ID: "read"},
		Resource:  EntityUID{Type: "Document", ID: "public"},
		Context:   Record{"time": String("office_hours")},
	}
	testId := "conditional-access"
	execTest(t, testId, request)
}

func TestPrincipalMismatch(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "valid"},
		Action:    EntityUID{Type: "Action", ID: "access"},
		Resource:  EntityUID{Type: "Service", ID: "restricted"},
		Context:   nil,
	}
	testId := "principal-mismatch"
	execTest(t, testId, request)
}

func TestEmptyPolicies(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: ""},
		Action:    EntityUID{Type: "Action", ID: ""},
		Resource:  EntityUID{Type: "Service", ID: ""},
		Context:   nil,
	}
	testId := "empty-policies"
	execTest(t, testId, request)
}

func TestLocationBasedCondition(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "userX"},
		Action:    EntityUID{Type: "Action", ID: "modify"},
		Resource:  EntityUID{Type: "System", ID: "core"},
		Context:   Record{"location": String("headquarters")},
	}
	testId := "location-based-condition"
	execTest(t, testId, request)
}

func TestInvalidEffect(t *testing.T) {
	request := Request{
		Principal: EntityUID{Type: "User", ID: "userY"},
		Action:    EntityUID{Type: "Action", ID: "access"},
		Resource:  EntityUID{Type: "Area", ID: "restricted"},
		Context:   nil,
	}
	testId := "invalid-effect"
	execTest(t, testId, request)
}
