package application

import (
	"strings"

	"github.com/coopa11y/DraftMeld/backend/internal/domain/ranking"
)

var nflTeamAliases = map[string]string{
	"ari": "ARI", "arizona": "ARI", "arizonacardinals": "ARI", "cardinals": "ARI",
	"atl": "ATL", "atlanta": "ATL", "atlantafalcons": "ATL", "falcons": "ATL",
	"bal": "BAL", "baltimore": "BAL", "baltimoreravens": "BAL", "ravens": "BAL",
	"buf": "BUF", "buffalo": "BUF", "buffalobills": "BUF", "bills": "BUF",
	"car": "CAR", "carolina": "CAR", "carolinapanthers": "CAR", "panthers": "CAR",
	"chi": "CHI", "chicago": "CHI", "chicagobears": "CHI", "bears": "CHI",
	"cin": "CIN", "cincinnati": "CIN", "cincinnatibengals": "CIN", "bengals": "CIN",
	"cle": "CLE", "cleveland": "CLE", "clevelandbrowns": "CLE", "browns": "CLE",
	"dal": "DAL", "dallas": "DAL", "dallascowboys": "DAL", "cowboys": "DAL",
	"den": "DEN", "denver": "DEN", "denverbroncos": "DEN", "broncos": "DEN",
	"det": "DET", "detroit": "DET", "detroitlions": "DET", "lions": "DET",
	"gb": "GB", "gbp": "GB", "greenbay": "GB", "greenbaypackers": "GB", "packers": "GB",
	"hou": "HOU", "houston": "HOU", "houstontexans": "HOU", "texans": "HOU",
	"ind": "IND", "indianapolis": "IND", "indianapoliscolts": "IND", "colts": "IND",
	"jax": "JAX", "jac": "JAX", "jacksonville": "JAX", "jacksonvillejaguars": "JAX", "jaguars": "JAX",
	"kc": "KC", "kcc": "KC", "kansascity": "KC", "kansascitychiefs": "KC", "chiefs": "KC",
	"lv": "LV", "lvr": "LV", "oak": "LV", "lasvegas": "LV", "lasvegasraiders": "LV", "raiders": "LV",
	"lac": "LAC", "sd": "LAC", "sdg": "LAC", "losangeleschargers": "LAC", "chargers": "LAC",
	"lar": "LAR", "stl": "LAR", "losangelesrams": "LAR", "rams": "LAR",
	"mia": "MIA", "miami": "MIA", "miamidolphins": "MIA", "dolphins": "MIA",
	"min": "MIN", "minnesota": "MIN", "minnesotavikings": "MIN", "vikings": "MIN",
	"ne": "NE", "nep": "NE", "newengland": "NE", "newenglandpatriots": "NE", "patriots": "NE",
	"no": "NO", "nos": "NO", "neworleans": "NO", "neworleanssaints": "NO", "saints": "NO",
	"nyg": "NYG", "newyorkgiants": "NYG", "giants": "NYG",
	"nyj": "NYJ", "newyorkjets": "NYJ", "jets": "NYJ",
	"phi": "PHI", "philadelphia": "PHI", "philadelphiaeagles": "PHI", "eagles": "PHI",
	"pit": "PIT", "pittsburgh": "PIT", "pittsburghsteelers": "PIT", "steelers": "PIT",
	"sea": "SEA", "seattle": "SEA", "seattleseahawks": "SEA", "seahawks": "SEA",
	"sf": "SF", "sfo": "SF", "sanfrancisco": "SF", "sanfrancisco49ers": "SF", "49ers": "SF",
	"tb": "TB", "tbb": "TB", "tampa": "TB", "tampabay": "TB", "tampabaybuccaneers": "TB", "buccaneers": "TB", "bucs": "TB",
	"ten": "TEN", "tennessee": "TEN", "tennesseetitans": "TEN", "titans": "TEN",
	"was": "WAS", "wsh": "WAS", "washington": "WAS", "washingtoncommanders": "WAS", "commanders": "WAS", "washingtonfootballteam": "WAS",
}

func normalizePosition(position string) string {
	normalized := strings.ToUpper(strings.TrimSpace(position))
	switch normalized {
	case "D/ST", "DEF", "DEFENSE":
		return "DST"
	default:
		return normalized
	}
}

func canonicalRankingKey(name, position, team string) string {
	if normalizePosition(position) != "DST" {
		return normalizePlayerKey(name)
	}
	if code := canonicalNFLTeam(team); code != "" {
		return "dst" + strings.ToLower(code)
	}
	if code := canonicalNFLTeam(name); code != "" {
		return "dst" + strings.ToLower(code)
	}
	return normalizePlayerKey(name)
}

func canonicalNFLTeam(value string) string {
	key := normalizePlayerKey(value)
	for _, suffix := range []string{"defensivespecialteams", "defense", "dst", "def"} {
		key = strings.TrimSuffix(key, suffix)
	}
	return nflTeamAliases[key]
}

func canonicalizeRankingRecord(record ranking.Record) ranking.Record {
	record.Position = normalizePosition(record.Position)
	if record.Position == "DST" {
		if record.PlayerKey == "" || !strings.HasPrefix(record.PlayerKey, "player-") {
			record.PlayerKey = canonicalRankingKey(record.Name, record.Position, record.Team)
		}
		if team := canonicalNFLTeam(record.Team); team != "" {
			record.Team = team
		} else if team = canonicalNFLTeam(record.Name); team != "" {
			record.Team = team
		}
	}
	if record.PlayerKey == "" {
		record.PlayerKey = canonicalRankingKey(record.Name, record.Position, record.Team)
	}
	return record
}
