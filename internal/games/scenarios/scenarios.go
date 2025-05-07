package scenarios

import (
	"embed"
	"fmt"
)

//go:embed *.txt
var fs embed.FS

func FromID(id string) (string, error) {
	data, err := fs.ReadFile(id + ".txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

type (
	violenceLevel uint8
	duration      uint8
)

func ViolenceLevel(i uint8) violenceLevel {
	if i < 0 {
		fmt.Printf("violence level %d < 0, setting to 0", i)
		i = 0
	} else if i > 3 {
		fmt.Printf("violence level %d > 3, setting to 3", i)
		i = 3
	}
	return violenceLevel(i)
}

func (vl violenceLevel) String() string {
	switch vl {
	case 0:
		return "Gar nicht gewalttätig"
	case 1:
		return "Leicht gewalttätig"
	case 2:
		return "Gewalttätig und grausam"
	case 3:
		return "Übertrieben gewalttätig, grausam und unangenehm"
	default:
		return "Unbekannt"
	}
}

func Duration(i uint8) duration {
	if i < 0 {
		fmt.Printf("duration %d < 0, setting to 0", i)
		i = 0
	} else if i > 2 {
		fmt.Printf("duration %d > 3, setting to 3", i)
		i = 2
	}
	return duration(i)
}

func (d duration) String() string {
	switch d {
	case 0:
		return "Sehr kurz (30-60 Minuten)"
	case 1:
		return "Kurz (2-4 Stunden)"
	case 2:
		return "Lang (4-8 Stunden)"
	default:
		return "Unbekannt"
	}
}
