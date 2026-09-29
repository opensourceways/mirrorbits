// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package rpc

import (
	"testing"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
)

func tsPtr(s int64, n int32) *timestamp.Timestamp {
	return &timestamp.Timestamp{Seconds: s, Nanos: n}
}

func TestScanMirrorRequest_Method_Enum(t *testing.T) {
	m := ScanMirrorRequest_FTP
	if m.Enum() == nil {
		t.Fatalf("Enum() should not return nil")
	}
	if m.String() != "FTP" {
		t.Fatalf("expected FTP, got %s", m.String())
	}
	_ = ScanMirrorRequest_Method(99).String()
	if ScanMirrorRequest_ALL.String() != "ALL" {
		t.Fatalf("expected ALL")
	}
	if ScanMirrorRequest_RSYNC.String() != "RSYNC" {
		t.Fatalf("expected RSYNC")
	}
	if ScanMirrorRequest_FTP.Descriptor() == nil {
		t.Fatalf("Descriptor() should not return nil")
	}
	if ScanMirrorRequest_FTP.Type() == nil {
		t.Fatalf("Type() should not return nil")
	}
	if ScanMirrorRequest_FTP.Number() != 1 {
		t.Fatalf("expected 1, got %d", ScanMirrorRequest_FTP.Number())
	}
	if ScanMirrorRequest_RSYNC.Number() != 2 {
		t.Fatalf("expected 2, got %d", ScanMirrorRequest_RSYNC.Number())
	}
	if ScanMirrorRequest_ALL.Number() != 0 {
		t.Fatalf("expected 0, got %d", ScanMirrorRequest_ALL.Number())
	}
	if _, idx := ScanMirrorRequest_FTP.EnumDescriptor(); len(idx) == 0 {
		t.Fatalf("EnumDescriptor() should return non-empty index")
	}
}

func TestVersionReply_Getters(t *testing.T) {
	x := &VersionReply{
		Version: "1.0", Build: "b1", GoVersion: "go1.20", OS: "linux", Arch: "amd64", GoMaxProcs: 8,
	}
	if x.GetVersion() != "1.0" {
		t.Fatalf("version mismatch")
	}
	if x.GetBuild() != "b1" {
		t.Fatalf("build mismatch")
	}
	if x.GetGoVersion() != "go1.20" {
		t.Fatalf("goversion mismatch")
	}
	if x.GetOS() != "linux" {
		t.Fatalf("os mismatch")
	}
	if x.GetArch() != "amd64" {
		t.Fatalf("arch mismatch")
	}
	if x.GetGoMaxProcs() != 8 {
		t.Fatalf("gomaxprocs mismatch")
	}
	var np *VersionReply
	if np.GetVersion() != "" || np.GetBuild() != "" || np.GetGoVersion() != "" ||
		np.GetOS() != "" || np.GetArch() != "" || np.GetGoMaxProcs() != 0 {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestVersionReply_ProtobufMethods(t *testing.T) {
	x := &VersionReply{Version: "v"}
	x.Reset()
	if x.GetVersion() != "" {
		t.Fatalf("Reset should clear fields")
	}
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect() should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor() should not be nil")
	}
	var np *VersionReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMatchRequest_Getters(t *testing.T) {
	x := &MatchRequest{Pattern: "*.iso"}
	if x.GetPattern() != "*.iso" {
		t.Fatalf("pattern mismatch")
	}
	var np *MatchRequest
	if np.GetPattern() != "" {
		t.Fatalf("nil getter should return empty")
	}
}

func TestMatchRequest_ProtobufMethods(t *testing.T) {
	x := &MatchRequest{Pattern: "p"}
	x.Reset()
	if x.GetPattern() != "" {
		t.Fatalf("Reset should clear")
	}
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *MatchRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMirror_Getters(t *testing.T) {
	x := &Mirror{
		ID: 1, Name: "m", HttpURL: "http://h", RsyncURL: "rsync://r", FtpURL: "ftp://f",
		SponsorName: "sn", SponsorURL: "su", SponsorLogoURL: "sl", AdminName: "an",
		AdminEmail: "ae", CustomData: "cd", ContinentOnly: true, CountryOnly: true,
		ASOnly: true, Score: 5, Latitude: 1.5, Longitude: 2.5, ContinentCode: "EU",
		CountryCodes: "FR", ExcludedCountryCodes: "DE", Asnum: 100, Comment: "c",
		Enabled: true, Up: true, ExcludeReason: "reason", StateSince: tsPtr(1, 1),
		AllowRedirects: 3, LastSync: tsPtr(2, 2), LastSuccessfulSync: tsPtr(3, 3),
		LastModTime: tsPtr(4, 4), Country: "FR", NetworkBandwidth: 1000,
	}
	if x.GetID() != 1 {
		t.Fatalf("ID mismatch")
	}
	if x.GetName() != "m" {
		t.Fatalf("Name mismatch")
	}
	if x.GetHttpURL() != "http://h" {
		t.Fatalf("HttpURL mismatch")
	}
	if x.GetRsyncURL() != "rsync://r" {
		t.Fatalf("RsyncURL mismatch")
	}
	if x.GetFtpURL() != "ftp://f" {
		t.Fatalf("FtpURL mismatch")
	}
	if x.GetSponsorName() != "sn" {
		t.Fatalf("SponsorName mismatch")
	}
	if x.GetSponsorURL() != "su" {
		t.Fatalf("SponsorURL mismatch")
	}
	if x.GetSponsorLogoURL() != "sl" {
		t.Fatalf("SponsorLogoURL mismatch")
	}
	if x.GetAdminName() != "an" {
		t.Fatalf("AdminName mismatch")
	}
	if x.GetAdminEmail() != "ae" {
		t.Fatalf("AdminEmail mismatch")
	}
	if x.GetCustomData() != "cd" {
		t.Fatalf("CustomData mismatch")
	}
	if !x.GetContinentOnly() {
		t.Fatalf("ContinentOnly mismatch")
	}
	if !x.GetCountryOnly() {
		t.Fatalf("CountryOnly mismatch")
	}
	if !x.GetASOnly() {
		t.Fatalf("ASOnly mismatch")
	}
	if x.GetScore() != 5 {
		t.Fatalf("Score mismatch")
	}
	if x.GetLatitude() != 1.5 {
		t.Fatalf("Latitude mismatch")
	}
	if x.GetLongitude() != 2.5 {
		t.Fatalf("Longitude mismatch")
	}
	if x.GetContinentCode() != "EU" {
		t.Fatalf("ContinentCode mismatch")
	}
	if x.GetCountryCodes() != "FR" {
		t.Fatalf("CountryCodes mismatch")
	}
	if x.GetExcludedCountryCodes() != "DE" {
		t.Fatalf("ExcludedCountryCodes mismatch")
	}
	if x.GetAsnum() != 100 {
		t.Fatalf("Asnum mismatch")
	}
	if x.GetComment() != "c" {
		t.Fatalf("Comment mismatch")
	}
	if !x.GetEnabled() {
		t.Fatalf("Enabled mismatch")
	}
	if !x.GetUp() {
		t.Fatalf("Up mismatch")
	}
	if x.GetExcludeReason() != "reason" {
		t.Fatalf("ExcludeReason mismatch")
	}
	if x.GetStateSince() == nil {
		t.Fatalf("StateSince should not be nil")
	}
	if x.GetAllowRedirects() != 3 {
		t.Fatalf("AllowRedirects mismatch")
	}
	if x.GetLastSync() == nil {
		t.Fatalf("LastSync should not be nil")
	}
	if x.GetLastSuccessfulSync() == nil {
		t.Fatalf("LastSuccessfulSync should not be nil")
	}
	if x.GetLastModTime() == nil {
		t.Fatalf("LastModTime should not be nil")
	}
	if x.GetCountry() != "FR" {
		t.Fatalf("Country mismatch")
	}
	if x.GetNetworkBandwidth() != 1000 {
		t.Fatalf("NetworkBandwidth mismatch")
	}
	var np *Mirror
	if np.GetID() != 0 || np.GetName() != "" || np.GetHttpURL() != "" || np.GetStateSince() != nil {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestMirror_ProtobufMethods(t *testing.T) {
	x := &Mirror{Name: "test"}
	x.Reset()
	if x.GetName() != "" {
		t.Fatalf("Reset should clear")
	}
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *Mirror
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMirrorListReply_Getters(t *testing.T) {
	x := &MirrorListReply{Mirrors: []*Mirror{{ID: 1}}}
	if len(x.GetMirrors()) != 1 {
		t.Fatalf("mirrors count mismatch")
	}
	var np *MirrorListReply
	if np.GetMirrors() != nil {
		t.Fatalf("nil getter should return nil")
	}
}

func TestMirrorListReply_ProtobufMethods(t *testing.T) {
	x := &MirrorListReply{}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *MirrorListReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMirrorID_Getters(t *testing.T) {
	x := &MirrorID{ID: 5, Name: "id"}
	if x.GetID() != 5 {
		t.Fatalf("ID mismatch")
	}
	if x.GetName() != "id" {
		t.Fatalf("Name mismatch")
	}
	var np *MirrorID
	if np.GetID() != 0 || np.GetName() != "" {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestMirrorID_ProtobufMethods(t *testing.T) {
	x := &MirrorID{Name: "x"}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *MirrorID
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMatchReply_Getters(t *testing.T) {
	x := &MatchReply{Mirrors: []*MirrorID{{ID: 1}}}
	if len(x.GetMirrors()) != 1 {
		t.Fatalf("mirrors count mismatch")
	}
	var np *MatchReply
	if np.GetMirrors() != nil {
		t.Fatalf("nil getter should return nil")
	}
}

func TestMatchReply_ProtobufMethods(t *testing.T) {
	x := &MatchReply{}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *MatchReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestChangeStatusRequest_Getters(t *testing.T) {
	x := &ChangeStatusRequest{ID: 3, Enabled: true}
	if x.GetID() != 3 {
		t.Fatalf("ID mismatch")
	}
	if !x.GetEnabled() {
		t.Fatalf("Enabled mismatch")
	}
	var np *ChangeStatusRequest
	if np.GetID() != 0 || np.GetEnabled() != false {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestChangeStatusRequest_ProtobufMethods(t *testing.T) {
	x := &ChangeStatusRequest{ID: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *ChangeStatusRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestMirrorIDRequest_Getters(t *testing.T) {
	x := &MirrorIDRequest{ID: 7}
	if x.GetID() != 7 {
		t.Fatalf("ID mismatch")
	}
	var np *MirrorIDRequest
	if np.GetID() != 0 {
		t.Fatalf("nil getter should return zero")
	}
}

func TestMirrorIDRequest_ProtobufMethods(t *testing.T) {
	x := &MirrorIDRequest{ID: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *MirrorIDRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestAddMirrorReply_Getters(t *testing.T) {
	x := &AddMirrorReply{
		Latitude: 1.1, Longitude: 2.2, Country: "US", Continent: "NA",
		ASN: "AS1", Warnings: []string{"w1"}, NetworkBandwidth: 500,
	}
	if x.GetLatitude() != 1.1 {
		t.Fatalf("Latitude mismatch")
	}
	if x.GetLongitude() != 2.2 {
		t.Fatalf("Longitude mismatch")
	}
	if x.GetCountry() != "US" {
		t.Fatalf("Country mismatch")
	}
	if x.GetContinent() != "NA" {
		t.Fatalf("Continent mismatch")
	}
	if x.GetASN() != "AS1" {
		t.Fatalf("ASN mismatch")
	}
	if len(x.GetWarnings()) != 1 {
		t.Fatalf("Warnings count mismatch")
	}
	if x.GetNetworkBandwidth() != 500 {
		t.Fatalf("NetworkBandwidth mismatch")
	}
	var np *AddMirrorReply
	if np.GetLatitude() != 0 || np.GetWarnings() != nil || np.GetNetworkBandwidth() != 0 {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestAddMirrorReply_ProtobufMethods(t *testing.T) {
	x := &AddMirrorReply{Country: "x"}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *AddMirrorReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestUpdateMirrorReply_Getters(t *testing.T) {
	x := &UpdateMirrorReply{Diff: "diff"}
	if x.GetDiff() != "diff" {
		t.Fatalf("Diff mismatch")
	}
	var np *UpdateMirrorReply
	if np.GetDiff() != "" {
		t.Fatalf("nil getter should return empty")
	}
}

func TestUpdateMirrorReply_ProtobufMethods(t *testing.T) {
	x := &UpdateMirrorReply{Diff: "d"}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *UpdateMirrorReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestRefreshRepositoryRequest_Getters(t *testing.T) {
	x := &RefreshRepositoryRequest{Rehash: true}
	if !x.GetRehash() {
		t.Fatalf("Rehash mismatch")
	}
	var np *RefreshRepositoryRequest
	if np.GetRehash() != false {
		t.Fatalf("nil getter should return false")
	}
}

func TestRefreshRepositoryRequest_ProtobufMethods(t *testing.T) {
	x := &RefreshRepositoryRequest{Rehash: true}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *RefreshRepositoryRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestScanMirrorRequest_Getters(t *testing.T) {
	x := &ScanMirrorRequest{ID: 9, AutoEnable: true, Protocol: ScanMirrorRequest_RSYNC}
	if x.GetID() != 9 {
		t.Fatalf("ID mismatch")
	}
	if !x.GetAutoEnable() {
		t.Fatalf("AutoEnable mismatch")
	}
	if x.GetProtocol() != ScanMirrorRequest_RSYNC {
		t.Fatalf("Protocol mismatch")
	}
	var np *ScanMirrorRequest
	if np.GetID() != 0 || np.GetAutoEnable() != false || np.GetProtocol() != ScanMirrorRequest_ALL {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestScanMirrorRequest_ProtobufMethods(t *testing.T) {
	x := &ScanMirrorRequest{ID: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *ScanMirrorRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestScanMirrorReply_Getters(t *testing.T) {
	x := &ScanMirrorReply{
		Enabled: true, FilesIndexed: 10, KnownIndexed: 20, Removed: 5, TZOffsetMs: 3600000,
	}
	if !x.GetEnabled() {
		t.Fatalf("Enabled mismatch")
	}
	if x.GetFilesIndexed() != 10 {
		t.Fatalf("FilesIndexed mismatch")
	}
	if x.GetKnownIndexed() != 20 {
		t.Fatalf("KnownIndexed mismatch")
	}
	if x.GetRemoved() != 5 {
		t.Fatalf("Removed mismatch")
	}
	if x.GetTZOffsetMs() != 3600000 {
		t.Fatalf("TZOffsetMs mismatch")
	}
	var np *ScanMirrorReply
	if np.GetEnabled() != false || np.GetFilesIndexed() != 0 || np.GetTZOffsetMs() != 0 {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestScanMirrorReply_ProtobufMethods(t *testing.T) {
	x := &ScanMirrorReply{Enabled: true}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *ScanMirrorReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestStatsFileRequest_Getters(t *testing.T) {
	x := &StatsFileRequest{Pattern: "*.txt", DateStart: tsPtr(1, 1), DateEnd: tsPtr(2, 2)}
	if x.GetPattern() != "*.txt" {
		t.Fatalf("Pattern mismatch")
	}
	if x.GetDateStart() == nil {
		t.Fatalf("DateStart should not be nil")
	}
	if x.GetDateEnd() == nil {
		t.Fatalf("DateEnd should not be nil")
	}
	var np *StatsFileRequest
	if np.GetPattern() != "" || np.GetDateStart() != nil || np.GetDateEnd() != nil {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestStatsFileRequest_ProtobufMethods(t *testing.T) {
	x := &StatsFileRequest{Pattern: "p"}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *StatsFileRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestStatsFileReply_Getters(t *testing.T) {
	x := &StatsFileReply{Files: map[string]int64{"a": 1, "b": 2}}
	if len(x.GetFiles()) != 2 {
		t.Fatalf("files count mismatch")
	}
	var np *StatsFileReply
	if np.GetFiles() != nil {
		t.Fatalf("nil getter should return nil")
	}
}

func TestStatsFileReply_ProtobufMethods(t *testing.T) {
	x := &StatsFileReply{Files: map[string]int64{"a": 1}}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *StatsFileReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestStatsMirrorRequest_Getters(t *testing.T) {
	x := &StatsMirrorRequest{ID: 3, DateStart: tsPtr(1, 1), DateEnd: tsPtr(2, 2)}
	if x.GetID() != 3 {
		t.Fatalf("ID mismatch")
	}
	if x.GetDateStart() == nil {
		t.Fatalf("DateStart should not be nil")
	}
	if x.GetDateEnd() == nil {
		t.Fatalf("DateEnd should not be nil")
	}
	var np *StatsMirrorRequest
	if np.GetID() != 0 || np.GetDateStart() != nil || np.GetDateEnd() != nil {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestStatsMirrorRequest_ProtobufMethods(t *testing.T) {
	x := &StatsMirrorRequest{ID: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *StatsMirrorRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestStatsMirrorReply_Getters(t *testing.T) {
	x := &StatsMirrorReply{Mirror: &Mirror{ID: 1}, Requests: 100, Bytes: 1024}
	if x.GetMirror() == nil {
		t.Fatalf("Mirror should not be nil")
	}
	if x.GetRequests() != 100 {
		t.Fatalf("Requests mismatch")
	}
	if x.GetBytes() != 1024 {
		t.Fatalf("Bytes mismatch")
	}
	var np *StatsMirrorReply
	if np.GetMirror() != nil || np.GetRequests() != 0 || np.GetBytes() != 0 {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestStatsMirrorReply_ProtobufMethods(t *testing.T) {
	x := &StatsMirrorReply{Requests: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *StatsMirrorReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestGetMirrorLogsRequest_Getters(t *testing.T) {
	x := &GetMirrorLogsRequest{ID: 5, MaxResults: 10}
	if x.GetID() != 5 {
		t.Fatalf("ID mismatch")
	}
	if x.GetMaxResults() != 10 {
		t.Fatalf("MaxResults mismatch")
	}
	var np *GetMirrorLogsRequest
	if np.GetID() != 0 || np.GetMaxResults() != 0 {
		t.Fatalf("nil getters should return zero values")
	}
}

func TestGetMirrorLogsRequest_ProtobufMethods(t *testing.T) {
	x := &GetMirrorLogsRequest{ID: 1}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *GetMirrorLogsRequest
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestGetMirrorLogsReply_Getters(t *testing.T) {
	x := &GetMirrorLogsReply{Line: []string{"line1", "line2"}}
	if len(x.GetLine()) != 2 {
		t.Fatalf("line count mismatch")
	}
	var np *GetMirrorLogsReply
	if np.GetLine() != nil {
		t.Fatalf("nil getter should return nil")
	}
}

func TestGetMirrorLogsReply_ProtobufMethods(t *testing.T) {
	x := &GetMirrorLogsReply{Line: []string{"a"}}
	x.Reset()
	x.ProtoMessage()
	_ = x.String()
	if x.ProtoReflect() == nil {
		t.Fatalf("ProtoReflect should not be nil")
	}
	if d, _ := x.Descriptor(); d == nil {
		t.Fatalf("Descriptor should not be nil")
	}
	var np *GetMirrorLogsReply
	if np.ProtoReflect() == nil {
		t.Fatalf("nil ProtoReflect should not be nil")
	}
}

func TestFileRpcProtoRawDescGZIP(t *testing.T) {
	data := file_rpc_proto_rawDescGZIP()
	if len(data) == 0 {
		t.Fatalf("rawDescGZIP should not be empty")
	}
	if file_rpc_proto_rawDescGZIP() == nil {
		t.Fatalf("second call should not be nil")
	}
}
