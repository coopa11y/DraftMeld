package application

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseESPNPPRRankings(t *testing.T) {
	players := make([]string, 0, 101)
	for rankValue := 1; rankValue <= 100; rankValue++ {
		positionID, teamID := 2, 1
		name := fmt.Sprintf("Player %d", rankValue)
		if rankValue == 1 {
			name, positionID, teamID = "Jahmyr Gibbs", 2, 8
		}
		if rankValue == 100 {
			name, positionID, teamID = "Denver Broncos", 16, 7
		}
		players = append(players, fmt.Sprintf(`{"player":{"id":%d,"fullName":%q,"defaultPositionId":%d,"proTeamId":%d,"draftRanksByRankType":{"PPR":{"rank":%d}},"ownership":{"averageDraftPosition":%g}}}`, 1000+rankValue, name, positionID, teamID, rankValue, float64(rankValue)+0.5))
	}
	players = append(players, `{"player":{"id":9999,"fullName":"Unsupported","defaultPositionId":11,"proTeamId":1,"draftRanksByRankType":{"PPR":{"rank":101}}}}`)
	records, published, err := parseESPNPPR(strings.NewReader(`{"players":[` + strings.Join(players, ",") + `]}`))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 100 || records[0].Name != "Jahmyr Gibbs" || records[0].Team != "DET" || records[0].ProviderID != "1001" || records[0].ADP != 1.5 {
		t.Fatalf("unexpected ESPN records: %#v", records[:1])
	}
	if records[99].Position != "DST" || records[99].Team != "DEN" || records[99].PlayerKey != "dstden" || published == "" {
		t.Fatalf("unexpected ESPN defense or publication: %#v published=%q", records[99], published)
	}
}

func TestParseYahooStandardRankings(t *testing.T) {
	var input strings.Builder
	input.WriteString(`<h3>Top 200 Default Rankings - Standard</h3><ol>`)
	for rankValue := 1; rankValue <= 100; rankValue++ {
		name := fmt.Sprintf("Player %d", rankValue)
		if rankValue == 1 {
			name = "Josh Allen"
		}
		if rankValue == 100 {
			name = "Denver Broncos"
		}
		fmt.Fprintf(&input, `<li class="Listitem Phone-fz-lg">%d. %s</li>`, rankValue, name)
	}
	input.WriteString(`</ol>`)
	records, published, err := parseYahooStandard(strings.NewReader(input.String()))
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 100 || records[0].Name != "Josh Allen" || records[0].PlayerKey != "joshallen" || published == "" {
		t.Fatalf("unexpected Yahoo records: %#v published=%q", records[:1], published)
	}
	if records[99].Position != "DST" || records[99].Team != "DEN" || records[99].PlayerKey != "dstden" {
		t.Fatalf("Yahoo defense was not canonicalized: %#v", records[99])
	}
}

func TestPlatformRankingParsersRejectIncompleteResponses(t *testing.T) {
	if _, _, err := parseESPNPPR(strings.NewReader(`{"players":[]}`)); err == nil {
		t.Fatal("expected incomplete ESPN response to fail")
	}
	if _, _, err := parseYahooStandard(strings.NewReader(`<li class="Listitem Phone-fz-lg">1. Josh Allen</li>`)); err == nil {
		t.Fatal("expected incomplete Yahoo response to fail")
	}
}
