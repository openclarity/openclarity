import React, { useState } from 'react';
import { isEmpty } from 'lodash';
import { useMountMultiFetch } from 'hooks';
import DropdownSelect from 'components/DropdownSelect';
import Button from 'components/Button';
import { RUNTIME_SCAN_URL } from '../useProgressLoaderReducer';

import './scan-configuration.scss';

const ScanConfiguration = ({isInProgress, onStartScan, onStopScan, namespaces, scannedNamespaces}) => {
    const [selectedNamespaces, setSelectedNamespaces] = useState(scannedNamespaces);

    const onActionClick = () => {
        if (isInProgress) {
            onStopScan();

            return;
        }

        onStartScan(selectedNamespaces.map(({value}) => value));
    }

    return (
        <div className="scan-configuration-container">
            <div className="scan-select-title">Select the target namespaces to scan:</div>
            <DropdownSelect
                name="runtime-scan-select-namespaces"
                className="runtime-scan-select"
                value={selectedNamespaces}
                items={namespaces}
                onChange={setSelectedNamespaces}
                disabled={isInProgress}
                isMulti={true}
                clearable={true}
            />
            <Button onClick={onActionClick} disabled={isEmpty(selectedNamespaces)}>
                {`${isInProgress ? "Stop" : "Start"} scan`}
            </Button>
        </div>
    )
}

const ScanConfigWrapper = props => {
    const [{loading, data}] = useMountMultiFetch([
        {key: "namespaces", url: "namespaces"},
        {key: "scanProgress", url: `${RUNTIME_SCAN_URL}/progress`},
    ]);

    if (loading) {
        return null
    }

    const {namespaces, scanProgress} = data || {};
    const {scannedNamespaces} = scanProgress || {};
    
    return (
        <ScanConfiguration
            {...props}
            namespaces={(namespaces || []).map(({name}) => ({value: name, label: name}))}
            scannedNamespaces={(scannedNamespaces || []).map(namespcase => ({value: namespcase, label: namespcase}))}
        />
    )
}

export default ScanConfigWrapper;