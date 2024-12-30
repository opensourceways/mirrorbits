// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package http

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"

	. "github.com/opensourceways/mirrorbits/config"
	"github.com/opensourceways/mirrorbits/filesystem"
	"github.com/opensourceways/mirrorbits/mirrors"
	"github.com/opensourceways/mirrorbits/network"
	"github.com/opensourceways/mirrorbits/utils"
)

type mirrorSelection interface {
	// Selection must return an ordered list of selected mirror,
	// a list of rejected mirrors and and an error code.
	Selection(*Context, *filesystem.FileInfo, network.GeoIPRecord, mirrors.Mirrors, *Configuration) (mirrors.Mirrors, mirrors.Mirrors, error)
}

// DefaultEngine is the default algorithm used for mirror selection
type DefaultEngine struct{}

// Selection returns an ordered list of selected mirror, a list of rejected mirrors and and an error code
func (h DefaultEngine) Selection(ctx *Context, fileInfo *filesystem.FileInfo, clientInfo network.GeoIPRecord, pMirrors mirrors.Mirrors, cnf *Configuration) (mlist mirrors.Mirrors, excluded mirrors.Mirrors, err error) {
	mlist = pMirrors
	// Filter
	safeIndex := 0
	excluded = make([]mirrors.Mirror, 0, len(mlist))
	var closestMirror float32
	var farthestMirror float32
	for i, m := range mlist {
		// Does it support http? Is it well formated?
		if !strings.HasPrefix(m.HttpURL, "http://") && !strings.HasPrefix(m.HttpURL, "https://") {
			m.ExcludeReason = "Invalid URL"
			goto discard
		}
		// Is it enabled?
		if !m.Enabled {
			m.ExcludeReason = "Disabled"
			goto discard
		}
		// Is it up?
		if !m.Up {
			if m.ExcludeReason == "" {
				m.ExcludeReason = "Down"
			}
			goto discard
		}
		if cnf.SchemaStrictMatch {
			if ctx.SecureOption() == WITHTLS && !m.IsHTTPS() {
				m.ExcludeReason = "Not HTTPS"
				goto discard
			}
			if ctx.SecureOption() == WITHOUTTLS && m.IsHTTPS() {
				m.ExcludeReason = "Not HTTP"
				goto discard
			}
		}
		// Is it the same size / modtime as source?
		if m.FileInfo != nil {
			if m.FileInfo.Size != fileInfo.Size {
				m.ExcludeReason = "File size mismatch"
				goto discard
			}
			if !m.FileInfo.ModTime.IsZero() {
				mModTime := m.FileInfo.ModTime
				if cnf.FixTimezoneOffsets {
					mModTime = mModTime.Add(time.Duration(m.TZOffset) * time.Millisecond)
				}
				mModTime = mModTime.Truncate(m.LastSuccessfulSyncPrecision.Duration())
				lModTime := fileInfo.ModTime.Truncate(m.LastSuccessfulSyncPrecision.Duration())
				if !mModTime.Equal(lModTime) {
					m.ExcludeReason = fmt.Sprintf("Mod time mismatch (diff: %s)", lModTime.Sub(mModTime))
					goto discard
				}
			}
		}
		// Is it configured to serve its continent only?
		if m.ContinentOnly {
			if !clientInfo.IsValid() || clientInfo.ContinentCode != m.ContinentCode {
				m.ExcludeReason = "Continent only"
				goto discard
			}
		}
		// Is it configured to serve its country only?
		if m.CountryOnly {
			if !clientInfo.IsValid() || !utils.IsInSlice(clientInfo.CountryCode, m.CountryFields) {
				m.ExcludeReason = "Country only"
				goto discard
			}
		}
		// Is it in the same AS number?
		if m.ASOnly {
			if !clientInfo.IsValid() || clientInfo.ASNum != m.Asnum {
				m.ExcludeReason = "AS only"
				goto discard
			}
		}
		// Is the user's country code allowed on this mirror?
		if clientInfo.IsValid() && utils.IsInSlice(clientInfo.CountryCode, m.ExcludedCountryFields) {
			m.ExcludeReason = "User's country restriction"
			goto discard
		}
		if safeIndex == 0 {
			closestMirror = m.Distance
		} else if closestMirror > m.Distance {
			closestMirror = m.Distance
		}
		if m.Distance > farthestMirror {
			farthestMirror = m.Distance
		}
		mlist[safeIndex] = mlist[i]
		safeIndex++
		continue
	discard:
		excluded = append(excluded, m)
	}

	// Reduce the slice to its new size
	mlist = mlist[:safeIndex]

	if !clientInfo.IsValid() {
		// Shuffle the list
		//XXX Should we use the fallbacks instead?
		for i := range mlist {
			j := rand.Intn(i + 1)
			mlist[i], mlist[j] = mlist[j], mlist[i]
		}

		// Shortcut
		if !ctx.IsMirrorlist() {
			// Reduce the number of mirrors to process
			mlist = mlist[:utils.Min(5, len(mlist))]
		}
		return
	}

	// We're not interested in divisions by zero
	if closestMirror == 0 {
		closestMirror = math.SmallestNonzeroFloat32
	}

	/* Weight distribution for random selection [Probabilistic weight] */

	// Compute score for each mirror and return the mirrors eligible for weight distribution.
	// This includes:
	// - mirrors found in a 1.5x (configurable) range from the closest mirror
	// - mirrors targeting the given country (as primary or secondary)
	// - mirrors being in the same AS number
	baseScore := int(farthestMirror)
	for i := 0; i < len(mlist); i++ {
		m := &mlist[i]
		var countryScore, netRateScore, distanceScore int
		distanceScore = baseScore - int(m.Distance) + 1
		netRateScore = m.Score
		if utils.IsPrimaryCountry(clientInfo, m.CountryFields) {
			countryScore = 1
		}
		m.ComputedScore[0] = countryScore
		m.ComputedScore[1] = netRateScore
		m.ComputedScore[2] = distanceScore

		log.Infof("mirror %s is chosen for request file %s, compute score %d, base score %d, country score %d, netRate score %d, distance score %d",
			m.Name, fileInfo.Path, m.ComputedScore, baseScore, countryScore, netRateScore, distanceScore)
	}

	// Sort mirrors by computed score
	sort.Sort(mirrors.ByComputedScore{Mirrors: mlist})

	return
}
