// Copyright (c) 2014-2019 Ludovic Faujet
// Licensed under the MIT license

package mirrors

import (
	"testing"

	"github.com/opensourceways/mirrorbits/network"
)

func TestByRank_Less_ASMatch(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 1234, Distance: 100},
			{ID: 2, Asnum: 5678, Distance: 50},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ASNum: 1234},
	}
	if !m.Less(0, 1) {
		t.Fatalf("Expected true when mirror i matches client AS")
	}
}

func TestByRank_Less_ASMatchJ(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 9999, Distance: 50},
			{ID: 2, Asnum: 1234, Distance: 100},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ASNum: 1234},
	}
	if m.Less(0, 1) {
		t.Fatalf("Expected false when mirror j matches client AS")
	}
}

func TestByRank_Less_CountryMatch(t *testing.T) {
	m1 := &Mirror{ID: 1, Asnum: 100, CountryCodes: "FR DE", Distance: 100}
	m1.Prepare()
	m2 := &Mirror{ID: 2, Asnum: 100, CountryCodes: "DE", Distance: 50}
	m2.Prepare()

	m := ByRank{
		Mirrors:    Mirrors{*m1, *m2},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ASNum: 100},
	}
	if !m.Less(0, 1) {
		t.Fatalf("Expected true when mirror i matches client country")
	}
}

func TestByRank_Less_CountryMatchJ(t *testing.T) {
	m1 := &Mirror{ID: 1, Asnum: 100, CountryCodes: "DE", Distance: 50}
	m1.Prepare()
	m2 := &Mirror{ID: 2, Asnum: 100, CountryCodes: "FR", Distance: 100}
	m2.Prepare()

	m := ByRank{
		Mirrors:    Mirrors{*m1, *m2},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ASNum: 100},
	}
	if m.Less(0, 1) {
		t.Fatalf("Expected false when mirror j matches client country")
	}
}

func TestByRank_Less_ContinentMatch(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 100, ContinentCode: "EU", Distance: 100},
			{ID: 2, Asnum: 100, ContinentCode: "AS", Distance: 50},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU", ASNum: 100},
	}
	if !m.Less(0, 1) {
		t.Fatalf("Expected true when mirror i matches client continent")
	}
}

func TestByRank_Less_ContinentMatchJ(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 100, ContinentCode: "AS", Distance: 50},
			{ID: 2, Asnum: 100, ContinentCode: "EU", Distance: 100},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU", ASNum: 100},
	}
	if m.Less(0, 1) {
		t.Fatalf("Expected false when mirror j matches client continent")
	}
}

func TestByRank_Less_DistanceFallback(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 100, ContinentCode: "EU", Distance: 50},
			{ID: 2, Asnum: 100, ContinentCode: "EU", Distance: 100},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU", ASNum: 100},
	}
	if !m.Less(0, 1) {
		t.Fatalf("Expected true when mirror i is closer")
	}
}

func TestByRank_Less_SameAS(t *testing.T) {
	m := ByRank{
		Mirrors: Mirrors{
			{ID: 1, Asnum: 100, ContinentCode: "EU", Distance: 50},
			{ID: 2, Asnum: 100, ContinentCode: "EU", Distance: 100},
		},
		ClientInfo: network.GeoIPRecord{CountryCode: "FR", ContinentCode: "EU", ASNum: 100},
	}
	if !m.Less(0, 1) {
		t.Fatalf("Expected true when same AS and mirror i is closer")
	}
}
