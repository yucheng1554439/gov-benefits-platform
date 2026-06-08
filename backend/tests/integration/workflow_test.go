//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

const foodAssistanceProgramID = "33333333-3333-3333-3333-333333333302"
const demoAgencyID = "22222222-2222-2222-2222-222222222201"

func TestHappyPath_ApplyToApproval(t *testing.T) {
	skipIfNoAPI(t)
	ctx := context.Background()
	c := newClient()

	if err := c.login(ctx, "worker1@dpss.lacounty.gov", "Password123!"); err != nil {
		t.Fatalf("worker login: %v", err)
	}

	citizen := newClient()
	email := fmt.Sprintf("test.citizen.%d@example.com", time.Now().UnixNano())
	regResp, err := citizen.do(ctx, http.MethodPost, "/auth/register", map[string]interface{}{
		"email":      email,
		"password":   "Password123!",
		"first_name": "Test",
		"last_name":  "Citizen",
		"agency_id":  demoAgencyID,
	}, nil)
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if regResp.StatusCode != http.StatusCreated && regResp.StatusCode != http.StatusOK {
		t.Fatalf("register status: %d", regResp.StatusCode)
	}
	regResp.Body.Close()

	if err := citizen.login(ctx, email, "Password123!"); err != nil {
		t.Fatalf("citizen login: %v", err)
	}
	citizen.agency = demoAgencyID

	var createdCase struct {
		ID string `json:"id"`
	}
	appResp, err := citizen.do(ctx, http.MethodPost, "/applications", map[string]interface{}{
		"agency_id":         demoAgencyID,
		"program_id":        foodAssistanceProgramID,
		"household_size":    3,
		"annual_income":     22000,
		"employment_status": "employed_part_time",
		"zip_code":          "90001",
		"census_tract":      "6037400100",
		"form_data":         map[string]interface{}{},
	}, &createdCase)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if appResp.StatusCode != http.StatusCreated {
		t.Fatalf("apply status: %d", appResp.StatusCode)
	}
	caseID := createdCase.ID
	if caseID == "" {
		t.Fatal("expected case id")
	}

	transitions := []string{"under_review", "eligibility_review", "approved"}
	for _, status := range transitions {
		resp, err := c.do(ctx, http.MethodPatch, "/cases/"+caseID+"/status", map[string]string{
			"to_status": status,
		}, nil)
		if err != nil {
			t.Fatalf("transition to %s: %v", status, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("transition to %s status: %d", status, resp.StatusCode)
		}
		resp.Body.Close()
	}

	eligResp, err := c.do(ctx, http.MethodPost, "/cases/"+caseID+"/eligibility/evaluate", nil, nil)
	if err != nil {
		t.Fatalf("eligibility evaluate: %v", err)
	}
	if eligResp.StatusCode != http.StatusOK && eligResp.StatusCode != http.StatusCreated {
		t.Fatalf("eligibility evaluate status: %d", eligResp.StatusCode)
	}
	eligResp.Body.Close()

	benResp, err := c.do(ctx, http.MethodPost, "/cases/"+caseID+"/benefit/calculate", nil, nil)
	if err != nil {
		t.Fatalf("benefit calc: %v", err)
	}
	if benResp.StatusCode != http.StatusOK && benResp.StatusCode != http.StatusCreated {
		t.Fatalf("benefit calc status: %d", benResp.StatusCode)
	}
	benResp.Body.Close()

	var letter struct {
		ID string `json:"id"`
	}
	letterResp, err := c.do(ctx, http.MethodPost, "/cases/"+caseID+"/letters", map[string]string{
		"letter_type": "approval_notice",
	}, &letter)
	if err != nil {
		t.Fatalf("generate letter: %v", err)
	}
	if letterResp.StatusCode != http.StatusCreated {
		t.Fatalf("generate letter status: %d", letterResp.StatusCode)
	}
	if letter.ID == "" {
		t.Fatal("expected letter id")
	}
	letterResp.Body.Close()

	downloadResp, err := c.do(ctx, http.MethodGet, "/letters/"+letter.ID+"/download", nil, nil)
	if err != nil {
		t.Fatalf("download letter: %v", err)
	}
	if downloadResp.StatusCode != http.StatusOK {
		t.Fatalf("download letter status: %d", downloadResp.StatusCode)
	}
	downloadResp.Body.Close()
}

func TestAppealPath_DeniedToApproved(t *testing.T) {
	skipIfNoAPI(t)
	ctx := context.Background()

	supervisor := newClient()
	if err := supervisor.login(ctx, "supervisor1@dpss.lacounty.gov", "Password123!"); err != nil {
		t.Fatalf("supervisor login: %v", err)
	}
	supervisor.agency = demoAgencyID

	citizen := newClient()
	if err := citizen.login(ctx, "citizen1@example.com", "Password123!"); err != nil {
		t.Fatalf("citizen login: %v", err)
	}
	citizen.agency = demoAgencyID

	worker := newClient()
	if err := worker.login(ctx, "worker1@dpss.lacounty.gov", "Password123!"); err != nil {
		t.Fatalf("worker login: %v", err)
	}
	worker.agency = demoAgencyID

	var createdCase struct {
		ID string `json:"id"`
	}
	appResp, err := citizen.do(ctx, http.MethodPost, "/applications", map[string]interface{}{
		"agency_id":         demoAgencyID,
		"program_id":        foodAssistanceProgramID,
		"household_size":    2,
		"annual_income":     22000,
		"employment_status": "employed_part_time",
		"zip_code":          "90001",
		"form_data":         map[string]interface{}{},
	}, &createdCase)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if appResp.StatusCode != http.StatusCreated {
		t.Fatalf("apply status: %d", appResp.StatusCode)
	}
	caseID := createdCase.ID

	for _, status := range []string{"under_review", "denied"} {
		resp, err := worker.do(ctx, http.MethodPatch, "/cases/"+caseID+"/status", map[string]string{"to_status": status}, nil)
		if err != nil {
			t.Fatalf("transition to %s: %v", status, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("transition to %s status: %d", status, resp.StatusCode)
		}
		resp.Body.Close()
	}

	appealResp, err := citizen.do(ctx, http.MethodPost, "/cases/"+caseID+"/appeal", map[string]string{
		"grounds": "Income was miscalculated. I have updated pay stubs.",
	}, nil)
	if err != nil {
		t.Fatalf("file appeal: %v", err)
	}
	if appealResp.StatusCode != http.StatusCreated {
		t.Fatalf("appeal filing status: %d", appealResp.StatusCode)
	}
	appealResp.Body.Close()

	listResp, err := supervisor.do(ctx, http.MethodGet, "/appeals", nil, nil)
	if err != nil {
		t.Fatalf("list appeals: %v", err)
	}
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("list appeals status: %d", listResp.StatusCode)
	}
	listResp.Body.Close()
}
