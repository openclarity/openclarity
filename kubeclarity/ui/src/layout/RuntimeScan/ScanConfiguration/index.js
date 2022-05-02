import React, { useState } from 'react';
import classnames from 'classnames';
import { useMountMultiFetch } from 'hooks';
import DropdownSelect from 'components/DropdownSelect';
import Button from 'components/Button';
import Icon, { ICON_NAMES } from 'components/Icon';
import InfoIcon from 'components/InfoIcon';
import { RUNTIME_SCAN_URL } from '../useProgressLoaderReducer';
import ScanOptionsForm from './ScanOptionsForm';

import './scan-configuration.scss';

const SELECT_INFO_MESSAGE = "Select specific namespaces to scan or leave empty to scan the whole K8s cluster";

const ScanConfiguration = ({isInProgress, onStartScan, onStopScan, namespaces, scannedNamespaces}) => {
    const [selectedNamespaces, setSelectedNamespaces] = useState(scannedNamespaces);

    const [showOptionsForm, setShowOptionsForm] = useState(false);

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
            <InfoIcon tooltipId="scan-config-info-tooltip" tooltipText={SELECT_INFO_MESSAGE} large />
            <DropdownSelect
                name="runtime-scan-select-namespaces"
                className="runtime-scan-select"
                value={selectedNamespaces}
                items={namespaces}
                onChange={setSelectedNamespaces}
                disabled={isInProgress}
                isMulti={true}
                clearable={true}
                placeholder={SELECT_INFO_MESSAGE}
            />
            <div className={classnames("scan-options-button", {disabled: isInProgress})} onClick={() => setShowOptionsForm(true)}>
                <Icon name={ICON_NAMES.COG} />
                <span>Options</span>
            </div>
            <Button className="scan-button" onClick={onActionClick}>{`${isInProgress ? "Stop" : "Start"} scan`}</Button>
            {showOptionsForm && <ScanOptionsForm onClose={() => setShowOptionsForm(false)} />}
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