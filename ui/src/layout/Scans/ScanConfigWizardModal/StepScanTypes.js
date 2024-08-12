import React from 'react';
import { CheckboxField, FieldLabel } from 'components/Form';

const StepScanTypes = () => {
    return (
        <div className="scan-config-scan-types-step">
            <FieldLabel>What would you like to scan for?</FieldLabel>
            <CheckboxField name="scanFamiliesConfig.sbom.enabled" title="SBOM" />
            <CheckboxField name="scanFamiliesConfig.vulnerabilities.enabled" title="Vulnerabilities" />
            <CheckboxField name="scanFamiliesConfig.malware.enabled" title="Malware" />
            <CheckboxField name="scanFamiliesConfig.rootkits.enabled" title="Rootkits" />
            <CheckboxField name="scanFamiliesConfig.secrets.enabled" title="Secrets" />
            <CheckboxField name="scanFamiliesConfig.misconfigurations.enabled" title="Misconfigurations" />
            <CheckboxField name="scanFamiliesConfig.exploits.enabled" title="Exploits" />
        </div>
    )
}

export default StepScanTypes;