import React from 'react';
import { CheckboxField, FieldLabel } from 'components/Form';

const StepScanTypes = () => {
    return (
        <div className="scan-config-scan-types-step">
            <FieldLabel>What would you like to scan for?</FieldLabel>
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.sbom.enabled" title="SBOM" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.vulnerabilities.enabled" title="Vulnerabilities" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.malware.enabled" title="Malware" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.rootkits.enabled" title="Rootkits" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.secrets.enabled" title="Secrets" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.misconfigurations.enabled" title="Misconfigurations" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.exploits.enabled" title="Exploits" />
            <CheckboxField name="scanTemplate.assetScanTemplate.scanFamiliesConfig.infoFinder.enabled" title="Info Finder" />
        </div>
    )
}

export default StepScanTypes;
