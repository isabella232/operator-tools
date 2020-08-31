// Copyright © 2020 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package templatereconciler

import (
	"emperror.dev/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"

	"github.com/banzaicloud/operator-tools/pkg/helm"
	"github.com/banzaicloud/operator-tools/pkg/reconciler"
	"github.com/banzaicloud/operator-tools/pkg/utils"
)

func orderedChartObjectsWithState(releaseData *ReleaseData) ([]runtime.Object, reconciler.DesiredState, error) {
	objects, err := chartObjects(releaseData)
	if err != nil {
		return nil, nil, err
	}

	utils.RuntimeObjects(objects).Sort(utils.InstallResourceOrder)

	return objects, reconciler.StatePresent, nil
}

func chartObjects(releaseData *ReleaseData) ([]runtime.Object, error) {
	chartDefaultValues, err := helm.GetDefaultValues(releaseData.Chart)
	if err != nil {
		return nil, errors.WrapIf(err, "could not get prometheus helm default values")
	}

	chartDefaultValuesYaml := helm.Strimap{}
	if err := yaml.Unmarshal(chartDefaultValues, &chartDefaultValuesYaml); err != nil {
		return nil, errors.WrapIf(err, "could not marshal default values")
	}

	objects, err := helm.Render(releaseData.Chart, helm.MergeMaps(chartDefaultValuesYaml, releaseData.Values), helm.ReleaseOptions{
		Name:      releaseData.ReleaseName,
		IsInstall: true,
		IsUpgrade: false,
		Namespace: releaseData.Namespace,
	}, releaseData.ChartName)
	if err != nil {
		return nil, errors.WrapIff(err, "could not render %s helm manifest objects", releaseData.ChartName)
	}

	return objects, nil
}