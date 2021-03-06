// Copyright 2020 Google LLC. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usbarmory

import (
	"crypto/sha256"
	"fmt"
)

const measurementDomainPrefix = "armory_mkii"

// ExpectedMeasurement returns the expected on-device measurement hash for the
// given firmware image.
//
// For the USB Armory, this is SHA256("armory_mkii"||img)
func ExpectedMeasurement(img []byte) ([]byte, error) {
	hasher := sha256.New()
	if _, err := hasher.Write([]byte(measurementDomainPrefix)); err != nil {
		return nil, fmt.Errorf("failed to write measurement domain prefix to hasher: %w", err)
	}
	if _, err := hasher.Write(img); err != nil {
		return nil, fmt.Errorf("failed to write firmware image to hasher: %w", err)
	}
	fwHash := hasher.Sum(nil)
	return fwHash[:], nil
}
