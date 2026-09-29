// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"testing"
)

func TestVersionReplyGetters(t *testing.T) {
	m := &VersionReply{
		Version:    "1.0",
		Build:      "abc",
		GoVersion:  "go1.20",
		OS:         "linux",
		Arch:       "amd64",
		GoMaxProcs: 4,
	}
	if m.GetVersion() != "1.0" {
		t.Fatal("Version mismatch")
	}
	if m.GetBuild() != "abc" {
		t.Fatal("Build mismatch")
	}
	if m.GetGoVersion() != "go1.20" {
		t.Fatal("GoVersion mismatch")
	}
	if m.GetOS() != "linux" {
		t.Fatal("OS mismatch")
	}
	if m.GetArch() != "amd64" {
		t.Fatal("Arch mismatch")
	}
	if m.GetGoMaxProcs() != 4 {
		t.Fatal("GoMaxProcs mismatch")
	}
	m.Reset()
	m.ProtoMessage()
	if m.GetVersion() != "" {
		t.Fatal("Expected empty after reset")
	}
}

func TestVersionReplyNilGetters(t *testing.T) {
	var m *VersionReply
	if m.GetVersion() != "" {
		t.Fatal("Expected empty")
	}
	if m.GetGoMaxProcs() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestMatchRequestGetters(t *testing.T) {
	m := &MatchRequest{Pattern: "test*"}
	if m.GetPattern() != "test*" {
		t.Fatal("Pattern mismatch")
	}
	m.Reset()
	if m.GetPattern() != "" {
		t.Fatal("Expected empty after reset")
	}
}

func TestMatchRequestNilGetters(t *testing.T) {
	var m *MatchRequest
	if m.GetPattern() != "" {
		t.Fatal("Expected empty")
	}
}

func TestMirrorGetters(t *testing.T) {
	m := &Mirror{
		ID:                    1,
		Name:                  "test",
		HttpURL:               "http://test.com",
		RsyncURL:              "rsync://test.com",
		FtpURL:                "ftp://test.com",
		SponsorName:           "sponsor",
		SponsorURL:            "http://sponsor.com",
		SponsorLogoURL:        "http://logo.com",
		AdminName:             "admin",
		AdminEmail:            "admin@test.com",
		CustomData:            "custom",
		ContinentOnly:         true,
		CountryOnly:           false,
		ASOnly:                true,
		Score:                 5,
		Latitude:              48.85,
		Longitude:             2.35,
		ContinentCode:         "EU",
		CountryCodes:          "FR DE",
		ExcludedCountryCodes:  "UK",
		Asnum:                 123,
		Comment:               "comment",
		Enabled:               true,
		Up:                    true,
		ExcludeReason:         "test",
		AllowRedirects:        1,
		Country:               "France",
		NetworkBandwidth:      100,
	}
	if m.GetID() != 1 {
		t.Fatal("ID mismatch")
	}
	if m.GetName() != "test" {
		t.Fatal("Name mismatch")
	}
	if m.GetHttpURL() != "http://test.com" {
		t.Fatal("HttpURL mismatch")
	}
	if m.GetRsyncURL() != "rsync://test.com" {
		t.Fatal("RsyncURL mismatch")
	}
	if m.GetFtpURL() != "ftp://test.com" {
		t.Fatal("FtpURL mismatch")
	}
	if m.GetSponsorName() != "sponsor" {
		t.Fatal("SponsorName mismatch")
	}
	if m.GetSponsorURL() != "http://sponsor.com" {
		t.Fatal("SponsorURL mismatch")
	}
	if m.GetSponsorLogoURL() != "http://logo.com" {
		t.Fatal("SponsorLogoURL mismatch")
	}
	if m.GetAdminName() != "admin" {
		t.Fatal("AdminName mismatch")
	}
	if m.GetAdminEmail() != "admin@test.com" {
		t.Fatal("AdminEmail mismatch")
	}
	if m.GetCustomData() != "custom" {
		t.Fatal("CustomData mismatch")
	}
	if !m.GetContinentOnly() {
		t.Fatal("ContinentOnly mismatch")
	}
	if m.GetCountryOnly() {
		t.Fatal("CountryOnly mismatch")
	}
	if !m.GetASOnly() {
		t.Fatal("ASOnly mismatch")
	}
	if m.GetScore() != 5 {
		t.Fatal("Score mismatch")
	}
	if m.GetLatitude() != 48.85 {
		t.Fatal("Latitude mismatch")
	}
	if m.GetLongitude() != 2.35 {
		t.Fatal("Longitude mismatch")
	}
	if m.GetContinentCode() != "EU" {
		t.Fatal("ContinentCode mismatch")
	}
	if m.GetCountryCodes() != "FR DE" {
		t.Fatal("CountryCodes mismatch")
	}
	if m.GetExcludedCountryCodes() != "UK" {
		t.Fatal("ExcludedCountryCodes mismatch")
	}
	if m.GetAsnum() != 123 {
		t.Fatal("Asnum mismatch")
	}
	if m.GetComment() != "comment" {
		t.Fatal("Comment mismatch")
	}
	if !m.GetEnabled() {
		t.Fatal("Enabled mismatch")
	}
	if !m.GetUp() {
		t.Fatal("Up mismatch")
	}
	if m.GetExcludeReason() != "test" {
		t.Fatal("ExcludeReason mismatch")
	}
	if m.GetAllowRedirects() != 1 {
		t.Fatal("AllowRedirects mismatch")
	}
	if m.GetCountry() != "France" {
		t.Fatal("Country mismatch")
	}
	if m.GetNetworkBandwidth() != 100 {
		t.Fatal("NetworkBandwidth mismatch")
	}
	if m.GetStateSince() != nil {
		t.Fatal("Expected nil StateSince")
	}
	if m.GetLastSync() != nil {
		t.Fatal("Expected nil LastSync")
	}
	if m.GetLastSuccessfulSync() != nil {
		t.Fatal("Expected nil LastSuccessfulSync")
	}
	if m.GetLastModTime() != nil {
		t.Fatal("Expected nil LastModTime")
	}
	m.Reset()
	m.ProtoMessage()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestMirrorNilGetters(t *testing.T) {
	var m *Mirror
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetName() != "" {
		t.Fatal("Expected empty")
	}
	if m.GetStateSince() != nil {
		t.Fatal("Expected nil")
	}
}

func TestMirrorListReplyGetters(t *testing.T) {
	m := &MirrorListReply{
		Mirrors: []*Mirror{{ID: 1}, {ID: 2}},
	}
	if len(m.GetMirrors()) != 2 {
		t.Fatalf("Expected 2 mirrors, got %d", len(m.GetMirrors()))
	}
	m.Reset()
	if len(m.GetMirrors()) != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestMirrorListReplyNilGetters(t *testing.T) {
	var m *MirrorListReply
	if m.GetMirrors() != nil {
		t.Fatal("Expected nil")
	}
}

func TestMirrorIDGetters(t *testing.T) {
	m := &MirrorID{ID: 5, Name: "test"}
	if m.GetID() != 5 {
		t.Fatal("ID mismatch")
	}
	if m.GetName() != "test" {
		t.Fatal("Name mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestMirrorIDNilGetters(t *testing.T) {
	var m *MirrorID
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetName() != "" {
		t.Fatal("Expected empty")
	}
}

func TestMatchReplyGetters(t *testing.T) {
	m := &MatchReply{
		Mirrors: []*MirrorID{{ID: 1}, {ID: 2}},
	}
	if len(m.GetMirrors()) != 2 {
		t.Fatalf("Expected 2, got %d", len(m.GetMirrors()))
	}
	m.Reset()
	if len(m.GetMirrors()) != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestMatchReplyNilGetters(t *testing.T) {
	var m *MatchReply
	if m.GetMirrors() != nil {
		t.Fatal("Expected nil")
	}
}

func TestChangeStatusRequestGetters(t *testing.T) {
	m := &ChangeStatusRequest{ID: 3, Enabled: true}
	if m.GetID() != 3 {
		t.Fatal("ID mismatch")
	}
	if !m.GetEnabled() {
		t.Fatal("Enabled mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestChangeStatusRequestNilGetters(t *testing.T) {
	var m *ChangeStatusRequest
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestMirrorIDRequestGetters(t *testing.T) {
	m := &MirrorIDRequest{ID: 7}
	if m.GetID() != 7 {
		t.Fatal("ID mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestMirrorIDRequestNilGetters(t *testing.T) {
	var m *MirrorIDRequest
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestAddMirrorReplyGetters(t *testing.T) {
	m := &AddMirrorReply{
		Latitude:         48.85,
		Longitude:        2.35,
		Country:          "France",
		Continent:        "EU",
		ASN:              "AS123",
		Warnings:         []string{"warn1", "warn2"},
		NetworkBandwidth: 100,
	}
	if m.GetLatitude() != 48.85 {
		t.Fatal("Latitude mismatch")
	}
	if m.GetLongitude() != 2.35 {
		t.Fatal("Longitude mismatch")
	}
	if m.GetCountry() != "France" {
		t.Fatal("Country mismatch")
	}
	if m.GetContinent() != "EU" {
		t.Fatal("Continent mismatch")
	}
	if m.GetASN() != "AS123" {
		t.Fatal("ASN mismatch")
	}
	if len(m.GetWarnings()) != 2 {
		t.Fatal("Warnings mismatch")
	}
	if m.GetNetworkBandwidth() != 100 {
		t.Fatal("NetworkBandwidth mismatch")
	}
	m.Reset()
	if m.GetLatitude() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestAddMirrorReplyNilGetters(t *testing.T) {
	var m *AddMirrorReply
	if m.GetLatitude() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetWarnings() != nil {
		t.Fatal("Expected nil")
	}
}

func TestUpdateMirrorReplyGetters(t *testing.T) {
	m := &UpdateMirrorReply{Diff: "test diff"}
	if m.GetDiff() != "test diff" {
		t.Fatal("Diff mismatch")
	}
	m.Reset()
	if m.GetDiff() != "" {
		t.Fatal("Expected empty after reset")
	}
}

func TestUpdateMirrorReplyNilGetters(t *testing.T) {
	var m *UpdateMirrorReply
	if m.GetDiff() != "" {
		t.Fatal("Expected empty")
	}
}

func TestRefreshRepositoryRequestGetters(t *testing.T) {
	m := &RefreshRepositoryRequest{Rehash: true}
	if !m.GetRehash() {
		t.Fatal("Rehash mismatch")
	}
	m.Reset()
	if m.GetRehash() {
		t.Fatal("Expected false after reset")
	}
}

func TestRefreshRepositoryRequestNilGetters(t *testing.T) {
	var m *RefreshRepositoryRequest
	if m.GetRehash() {
		t.Fatal("Expected false")
	}
}

func TestScanMirrorRequestGetters(t *testing.T) {
	m := &ScanMirrorRequest{ID: 10, AutoEnable: true, Protocol: ScanMirrorRequest_FTP}
	if m.GetID() != 10 {
		t.Fatal("ID mismatch")
	}
	if !m.GetAutoEnable() {
		t.Fatal("AutoEnable mismatch")
	}
	if m.GetProtocol() != ScanMirrorRequest_FTP {
		t.Fatal("Protocol mismatch")
	}
	m.Reset()
	if m.GetID() != 0 {
		t.Fatal("Expected 0 after reset")
	}
}

func TestScanMirrorRequestNilGetters(t *testing.T) {
	var m *ScanMirrorRequest
	if m.GetID() != 0 {
		t.Fatal("Expected 0")
	}
	if m.GetProtocol() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestScanMirrorReplyGetters(t *testing.T) {
	m := &ScanMirrorReply{Enabled: true, FilesIndexed: 42}
	if !m.GetEnabled() {
		t.Fatal("Enabled mismatch")
	}
	if m.GetFilesIndexed() != 42 {
		t.Fatal("FilesIndexed mismatch")
	}
	m.Reset()
	if m.GetEnabled() {
		t.Fatal("Expected false after reset")
	}
}

func TestScanMirrorReplyNilGetters(t *testing.T) {
	var m *ScanMirrorReply
	if m.GetEnabled() {
		t.Fatal("Expected false")
	}
	if m.GetFilesIndexed() != 0 {
		t.Fatal("Expected 0")
	}
}

func TestScanMirrorRequestMethodEnum(t *testing.T) {
	v := ScanMirrorRequest_FTP
	if v.String() == "" {
		t.Fatal("Expected non-empty string")
	}
	_ = v.Number()
}

func TestMessageStringMethods(t *testing.T) {
	mirrors := []interface{ String() string; Reset(); ProtoMessage() }{
		&VersionReply{},
		&MatchRequest{},
		&Mirror{},
		&MirrorListReply{},
		&MirrorID{},
		&MatchReply{},
		&ChangeStatusRequest{},
		&MirrorIDRequest{},
		&AddMirrorReply{},
		&UpdateMirrorReply{},
		&RefreshRepositoryRequest{},
		&ScanMirrorRequest{},
		&ScanMirrorReply{},
	}
	for _, v := range mirrors {
		_ = v.String()
		v.Reset()
		v.ProtoMessage()
	}
}
