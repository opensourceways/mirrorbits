// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package network

import (
	"net"
	"testing"
	"time"
)

type GeoIPMockCityTWN struct{}

func (g *GeoIPMockCityTWN) Lookup(ipAddress net.IP, result interface{}) error {
	var citydb CityDb
	citydb.City.Names.En = "Taipei"
	citydb.Country.Iso_Code = "TWN"
	citydb.Country.Names.En = "Taiwan"
	citydb.Continent.Code = "AS"
	citydb.Location.Latitude = 25.0330
	citydb.Location.Longitude = 121.5654
	CopyStruct(&citydb, result)
	return nil
}

type GeoIPMockCityHK struct{}

func (g *GeoIPMockCityHK) Lookup(ipAddress net.IP, result interface{}) error {
	var citydb CityDb
	citydb.City.Names.En = "Hong Kong"
	citydb.Country.Iso_Code = "HK"
	citydb.Country.Names.En = "Hong Kong"
	citydb.Continent.Code = "AS"
	citydb.Location.Latitude = 22.3193
	citydb.Location.Longitude = 114.1694
	CopyStruct(&citydb, result)
	return nil
}

type GeoIPMockCityMO struct{}

func (g *GeoIPMockCityMO) Lookup(ipAddress net.IP, result interface{}) error {
	var citydb CityDb
	citydb.City.Names.En = "Macau"
	citydb.Country.Iso_Code = "MO"
	citydb.Country.Names.En = "Macau"
	citydb.Continent.Code = "AS"
	citydb.Location.Latitude = 22.1987
	citydb.Location.Longitude = 113.5439
	CopyStruct(&citydb, result)
	return nil
}

type GeoIPMockCityNormal struct{}

func (g *GeoIPMockCityNormal) Lookup(ipAddress net.IP, result interface{}) error {
	var citydb CityDb
	citydb.City.Names.En = "Paris"
	citydb.Country.Iso_Code = "FR"
	citydb.Country.Names.En = "France"
	citydb.Continent.Code = "EU"
	citydb.Location.Latitude = 48.8567
	citydb.Location.Longitude = 2.3508
	CopyStruct(&citydb, result)
	return nil
}

func TestGeoIP_GetRecord_TWN(t *testing.T) {
	g := NewGeoIP()

	g.city = &geoipDB{
		filename: "city.mmdb",
		modTime:  time.Now(),
		db:       &GeoIPMockCityTWN{},
	}

	r := g.GetRecord("1.2.3.4")
	if r.CountryCode != "CN" {
		t.Fatalf("Expected CountryCode 'CN' for TWN, got %s", r.CountryCode)
	}
	if r.Country != "China" {
		t.Fatalf("Expected Country 'China' for TWN, got %s", r.Country)
	}
}

func TestGeoIP_GetRecord_HK(t *testing.T) {
	g := NewGeoIP()

	g.city = &geoipDB{
		filename: "city.mmdb",
		modTime:  time.Now(),
		db:       &GeoIPMockCityHK{},
	}

	r := g.GetRecord("1.2.3.4")
	if r.CountryCode != "CN" {
		t.Fatalf("Expected CountryCode 'CN' for HK, got %s", r.CountryCode)
	}
	if r.Country != "China" {
		t.Fatalf("Expected Country 'China' for HK, got %s", r.Country)
	}
}

func TestGeoIP_GetRecord_MO(t *testing.T) {
	g := NewGeoIP()

	g.city = &geoipDB{
		filename: "city.mmdb",
		modTime:  time.Now(),
		db:       &GeoIPMockCityMO{},
	}

	r := g.GetRecord("1.2.3.4")
	if r.CountryCode != "CN" {
		t.Fatalf("Expected CountryCode 'CN' for MO, got %s", r.CountryCode)
	}
	if r.Country != "China" {
		t.Fatalf("Expected Country 'China' for MO, got %s", r.Country)
	}
}

func TestGeoIP_GetRecord_Normal(t *testing.T) {
	g := NewGeoIP()

	g.city = &geoipDB{
		filename: "city.mmdb",
		modTime:  time.Now(),
		db:       &GeoIPMockCityNormal{},
	}
	g.asn = &geoipDB{
		filename: "asn.mmdb",
		modTime:  time.Now(),
		db:       &GeoIPMockASN{},
	}

	r := g.GetRecord("1.2.3.4")
	if r.CountryCode != "FR" {
		t.Fatalf("Expected CountryCode 'FR', got %s", r.CountryCode)
	}
	if r.Country != "France" {
		t.Fatalf("Expected Country 'France', got %s", r.Country)
	}
	if r.City != "Paris" {
		t.Fatalf("Expected City 'Paris', got %s", r.City)
	}
	if r.ContinentCode != "EU" {
		t.Fatalf("Expected ContinentCode 'EU', got %s", r.ContinentCode)
	}
	if r.ASNum != 42 {
		t.Fatalf("Expected ASNum 42, got %d", r.ASNum)
	}
}

func TestGeoIP_GetRecord_InvalidIP(t *testing.T) {
	g := NewGeoIP()

	r := g.GetRecord("invalid-ip")
	if r.CountryCode != "" {
		t.Fatalf("Expected empty CountryCode for invalid IP")
	}
}

func TestGeoIP_GetRecord_NoDB(t *testing.T) {
	g := NewGeoIP()

	r := g.GetRecord("127.0.0.1")
	if r.CountryCode != "" {
		t.Fatalf("Expected empty CountryCode when no DB loaded")
	}
}

func TestGeoIPError_Error(t *testing.T) {
	e := GeoIPError{}
	if e.Error() != "One or more GeoIP database could not be loaded" {
		t.Fatalf("Expected error message")
	}
}

func TestGeoIPError_IsFatal_AllFailed(t *testing.T) {
	e := GeoIPError{
		Errors:  []error{nil, nil},
		loaded:  2,
	}
	if !e.IsFatal() {
		t.Fatalf("Expected true when all databases failed")
	}
}

func TestGeoIPError_IsFatal_PartialFailure(t *testing.T) {
	e := GeoIPError{
		Errors:  []error{nil},
		loaded:  2,
	}
	if e.IsFatal() {
		t.Fatalf("Expected false when not all databases failed")
	}
}

func TestErrMultipleAddresses(t *testing.T) {
	if ErrMultipleAddresses == nil {
		t.Fatalf("ErrMultipleAddresses should not be nil")
	}
	if ErrMultipleAddresses.Error() != "the mirror has more than one IP address" {
		t.Fatalf("Unexpected error message: %s", ErrMultipleAddresses.Error())
	}
}

func TestExtractRemoteIP_Single(t *testing.T) {
	r := ExtractRemoteIP("192.168.0.1")
	if r != "192.168.0.1" {
		t.Fatalf("Expected '192.168.0.1', got %s", r)
	}
}

func TestExtractRemoteIP_Empty(t *testing.T) {
	r := ExtractRemoteIP("")
	if r != "" {
		t.Fatalf("Expected '', got %s", r)
	}
}
