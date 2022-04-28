// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
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

package database

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ObjectTree interface {
	SetApplication(app *Application, params *TransactionParams, shouldUpdatePackageVulnerabilities bool) error
	SetResource(resource *Resource, params *TransactionParams, shouldUpdatePackageVulnerabilities bool) error
}

type ObjectTreeHandler struct {
	db *gorm.DB
}

func (o *ObjectTreeHandler) SetApplication(app *Application, params *TransactionParams, shouldUpdatePackageVulnerabilities bool) error {
	// DFS update of the tree and set the relations as in the received application.
	err := o.db.WithContext(createContextWithValues(params)).Transaction(func(tx *gorm.DB) error {
		return o.updateApplication(tx, app, shouldUpdatePackageVulnerabilities)
	})
	if err != nil {
		return fmt.Errorf("failed to update application: %v", err)
	}

	return nil
}

func (o *ObjectTreeHandler) updateApplication(tx *gorm.DB, app *Application, shouldUpdatePackageVulnerabilities bool) error {
	for i := range app.Resources {
		if err := o.updateResource(tx, &app.Resources[i], shouldUpdatePackageVulnerabilities); err != nil {
			return fmt.Errorf("failed to update resource: %v", err)
		}

		// set it to nil, so it will not be inserted again during the higher level association replace.
		app.Resources[i].Packages = nil
		app.Resources[i].CISDockerBenchmarkChecks = nil
	}

	log.Tracef("Updating application=%+v", app)
	if err := tx.Omit("Resources").
		Save(app).Error; err != nil {
		return fmt.Errorf("failed to update application: %v", err)
	}

	// Update resources and application_resources
	log.Tracef("Updating resources and application_resources. resources=%+v", app.Resources)
	if err := tx.Model(&app).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Association("Resources").Replace(app.Resources); err != nil {
		return fmt.Errorf("failed to update application resources association: %v", err)
	}

	return nil
}

func (o *ObjectTreeHandler) SetResource(resource *Resource, params *TransactionParams, shouldUpdatePackageVulnerabilities bool) error {
	// DFS update of the tree and set the relations as in the received resource.
	err := o.db.WithContext(createContextWithValues(params)).Transaction(func(tx *gorm.DB) error {
		return o.updateResource(tx, resource, shouldUpdatePackageVulnerabilities)
	})
	if err != nil {
		return fmt.Errorf("failed to update resource: %v", err)
	}

	return nil
}

func (o *ObjectTreeHandler) updateResource(tx *gorm.DB, resource *Resource, shouldUpdatePackageVulnerabilities bool) error {
	for i := range resource.Packages {
		if err := o.updatePackage(tx, &resource.Packages[i], shouldUpdatePackageVulnerabilities); err != nil {
			return fmt.Errorf("failed to update package: %v", err)
		}

		// set it to nil, so it will not be inserted again during the higher level association replace.
		resource.Packages[i].Vulnerabilities = nil
	}
	//
	//for i := range resource.CISDockerBenchmarkChecks {
	//	if err := o.updateCISDockerBenchmarkCheck(tx, &resource.CISDockerBenchmarkChecks[i]); err != nil {
	//		return fmt.Errorf("failed to update cis docker benchmark check: %v", err)
	//	}
	//}

	log.Tracef("Updating resource=%+v", resource)
	if err := tx.Omit("Packages", "CISDockerBenchmarkChecks").
		Save(resource).Error; err != nil {
		return fmt.Errorf("failed to update resource: %v", err)
	}

	// Update packages and resource_packages
	log.Tracef("Updating packages and resource_packages. packages=%+v", resource.Packages)
	if err := tx.Model(resource).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Association("Packages").Replace(resource.Packages); err != nil {
		return fmt.Errorf("failed to update resource packages association: %v", err)
	}

	// Update cis_docker_benchmark_results
	log.Tracef("Updating cis_docker_benchmark_checks and resource_cis_docker_benchmark_checks. checks=%+v", resource.CISDockerBenchmarkChecks)
	if err := tx.Model(resource).
		Session(&gorm.Session{FullSaveAssociations: true}).
		Association("CISDockerBenchmarkChecks").Replace(resource.CISDockerBenchmarkChecks); err != nil {
		return fmt.Errorf("failed to update resource cis_docker_benchmark_checks association: %v", err)
	}

	return nil
}

func (o *ObjectTreeHandler) updatePackage(tx *gorm.DB, pkg *Package, shouldUpdatePackageVulnerabilities bool) error {
	// Update package
	log.Tracef("Updating package=%+v", pkg)
	if err := tx.Omit("Vulnerabilities").
		Save(pkg).Error; err != nil {
		return fmt.Errorf("failed to update package: %v", err)
	}

	if shouldUpdatePackageVulnerabilities {
		// Update vulnerabilities and package_vulnerabilities
		log.Tracef("Updating vulnerabilities and package_vulnerabilities. vulnerabilities=%+v", pkg.Vulnerabilities)
		if err := tx.Model(pkg).
			Session(&gorm.Session{FullSaveAssociations: true}).
			Association("Vulnerabilities").Replace(pkg.Vulnerabilities); err != nil {
			return fmt.Errorf("failed to update package vulnerabilities association: %v", err)
		}
	}

	return nil
}

//
//func (o *ObjectTreeHandler) updateCISDockerBenchmarkCheck(tx *gorm.DB, check *CISDockerBenchmarkCheck) error {
//	// Update cis docker benchmark check
//	log.Tracef("Updating cis docker benchmark check=%+v", check)
//	if err := tx.Save(check).Error; err != nil {
//		return fmt.Errorf("failed to update cis docker benchmark check: %v", err)
//	}
//
//	return nil
//}
