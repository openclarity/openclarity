import React from 'react';
import { TextField, CheckboxField, validators } from 'components/Form';

const StepAdvancedSettings = () => (
    <div className="scan-config-advanced-settings-step">
        <TextField
            name="scanTemplate.maxParallelScanners"
            label="Maximal number of instances to be scanned in parallel"
            type="number"
            tooltipText={(
                <div style={{width: "350px"}}>
                    The maximum number of scanners (per instance) that that can run in parallel for each scan
                </div>
            )}
            validate={validators.validateRequired}
        />
        <CheckboxField
            name="scanTemplate.assetScanTemplate.scannerInstanceCreationConfig.useSpotInstances"
            title="Spot instances required"
            tooltipText={(
                <div style={{width: "350px"}}>
                    <div>When selected, Spot instances will be used for the VM scanners to lower billing prices.</div>
                    <div>However, this can result in unexpected Asset Scan failures if a spot instance is terminated during a scan.</div>
                </div>
            )}
        />
    </div>
)

export default StepAdvancedSettings;
