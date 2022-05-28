package boc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testData struct {
	date    string
	year2   string
	year3   string
	year5   string
	wantErr bool
}

func TestSuccess(t *testing.T) {
	a := assert.New(t)
	b, err := NewBOCInterests()
	a.NoError(err)
	a.NotNil(b)

	tests := []testData{
		{
			date:    "2022-05-23",
			wantErr: true,
		},
		{
			date:  "2022-05-24",
			year2: "2.57",
			year3: "2.58",
			year5: "2.64",
		},

		{
			date:  "2022-05-25",
			year2: "2.53",
			year3: "2.54",
			year5: "2.60",
		},
		{
			date:  "2022-05-26",
			year2: "2.55",
			year3: "2.55",
			year5: "2.62",
		},
	}

	for _, tt := range tests {
		obs, err := b.GetObservationForDate(tt.date)
		if tt.wantErr {
			a.Error(err)
			a.Nil(obs)
			continue
		}
		a.Equal(tt.date, obs.D)
		a.Equal(tt.year2, obs.Yield2Year.V)
		a.Equal(tt.year3, obs.Yield3Year.V)
		a.Equal(tt.year5, obs.Yield5Year.V)

	}

}

func TestFormatDate(t *testing.T) {
	tests := []struct {
		name    string
		date    string
		want    string
		wantErr bool
	}{
		{
			name: "success",
			date: "1990/20/12",
			want: "1990-12-20",
		},
		{
			name: "success",
			date: "10-07-1990",
			want: "1990-07-10",
		},
		{
			name: "success",
			date: "25-5-2000",
			want: "2000-05-25",
		},
		{
			name: "success",
			date: "5-5-2095",
			want: "2095-05-05",
		},
		{
			name: "success",
			date: "2000-24-5",
			want: "2000-05-24",
		},
		{
			name: "success",
			date: "12-13-2020",
			want: "2020-12-13",
		},
		{
			name: "success",
			date: "1990\\05\\20",
			want: "1990-05-20",
		},
		{
			name: "success",
			date: "1990/05/01",
			want: "1990-05-01",
		},
		{
			name:    "error",
			date:    "19906-05-01",
			wantErr: true,
		},
		{
			name:    "error",
			date:    "1990/01/-1",
			wantErr: true,
		},
		{
			name:    "error",
			date:    "1990\\50\\01",
			wantErr: true,
		},
		{
			name:    "error",
			date:    "20/13/1990",
			wantErr: true,
		},

		{
			name:    "error",
			date:    "10/10/f902",
			wantErr: true,
		},
		{
			name:    "error",
			date:    "10/-1/1990",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FormatDate(tt.date)
			if (err != nil) != tt.wantErr {
				t.Errorf("FormatDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FormatDate() = %v, want %v", got, tt.want)
			}
		})
	}
}
