package application

import (
	"context"
	"strings"
	"testing"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/projection"
)

type projectionRepositoryStub struct{ records []projection.Record }

func (stub *projectionRepositoryStub) ReplaceProjections(_ context.Context, _ projection.SourceStatus, records []projection.Record) error {
	stub.records = append([]projection.Record(nil), records...)
	return nil
}
func (stub *projectionRepositoryStub) ProjectionRecords(context.Context) ([]projection.Record, error) {
	return stub.records, nil
}
func (stub *projectionRepositoryStub) ProjectionStatuses(context.Context) ([]projection.SourceStatus, error) {
	return []projection.SourceStatus{}, nil
}

func TestProjectionCSVAppliesLeagueScoring(t *testing.T) {
	repository := &projectionRepositoryStub{}
	service := NewProjectionService(repository)
	_, err := service.ImportCSV(t.Context(), "Analyst A", strings.NewReader("name,position,team,adp,byeWeek,reception,receivingYard,receivingTouchdown\nAda Runner,RB,ATL,20,8,50,800,8\n"))
	if err != nil {
		t.Fatal(err)
	}
	values, err := service.LeagueValues(t.Context(), map[string]float64{"reception": 1, "receivingYard": .1, "receivingTouchdown": 6})
	if err != nil {
		t.Fatal(err)
	}
	value := values["adarunner"]
	if value.ProjectedPoints != 178 || value.ADP != 20 || value.ByeWeek != 8 {
		t.Fatalf("unexpected league projection: %#v", value)
	}
}

func TestProjectionCSVRequiresCanonicalColumns(t *testing.T) {
	service := NewProjectionService(&projectionRepositoryStub{})
	if _, err := service.ImportCSV(t.Context(), "Broken", strings.NewReader("player,pos\nAda,RB\n")); err == nil {
		t.Fatal("expected missing-column error")
	}
}

func TestProjectionCSVAcceptsExplicitColumnMapping(t *testing.T) {
	repository := &projectionRepositoryStub{}
	service := NewProjectionService(repository)
	_, err := service.ImportCSV(t.Context(), "Mapped", strings.NewReader("Player Full Name,Pos,Tm,Rec Total\nAlex Rivers,RB,ATL,72\n"), map[string]string{
		"name": "Player Full Name", "position": "Pos", "team": "Tm", "reception": "Rec Total",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(repository.records) != 1 || repository.records[0].Stats["reception"] != 72 {
		t.Fatalf("mapped import failed: %#v", repository.records)
	}
}
