import React, { useState } from 'react';
import { useMountMultiFetch } from 'hooks';
import DropdownSelect from 'components/DropdownSelect';
import Button from 'components/Button';
import { ICON_NAMES } from 'components/Icon';
import InfoIcon from 'components/InfoIcon';
import { RUNTIME_SCAN_URL } from '../useProgressLoaderReducer';
import ScanOptionsForm from './ScanOptionsForm';
import ScheduleScanForm from './ScheduleScanForm';
import { ConfigButton, NAMESPACES_SELECT_TITLE, NAMESPACES_SELECT_INFO_MESSAGE } from './utils';

import './scan-configuration.scss';

const ScanConfiguration = ({isInProgress, onStartScan, onStopScan, namespaces, scannedNamespaces}) => {
    const [selectedNamespaces, setSelectedNamespaces] = useState(scannedNamespaces);

    const [showOptionsForm, setShowOptionsForm] = useState(false);
    const [showScheduleScanForm, setShowScheduleScanForm] = useState(false);

    const onActionClick = () => {
        if (isInProgress) {
            onStopScan();

            return;
        }

        onStartScan(selectedNamespaces.map(({value}) => value));
    }

    return (
        <div className="scan-configuration-container">
            <div className="on-demand-scan-configuration-container">
                <div className="scan-select-title">{`${NAMESPACES_SELECT_TITLE}:`}</div>
                <InfoIcon tooltipId="scan-config-info-tooltip" tooltipText={NAMESPACES_SELECT_INFO_MESSAGE} large />
                <DropdownSelect
                    name="runtime-scan-select-namespaces"
                    className="runtime-scan-select"
                    value={selectedNamespaces}
                    items={namespaces}
                    onChange={setSelectedNamespaces}
                    disabled={isInProgress}
                    isMulti={true}
                    clearable={true}
                    placeholder={NAMESPACES_SELECT_INFO_MESSAGE}
                />
                <ConfigButton icon={ICON_NAMES.COG} disabled={isInProgress} onClick={() => setShowOptionsForm(true)}>Options</ConfigButton>
                <Button className="scan-button" onClick={onActionClick}>{`${isInProgress ? "Stop" : "Start"} scan`}</Button>
                {showOptionsForm && <ScanOptionsForm onClose={() => setShowOptionsForm(false)} />}
            </div>
            <div className="scheduled-scan-configuration-container">
                <ConfigButton icon={ICON_NAMES.CLOCK} onClick={() => setShowScheduleScanForm(true)}>Schedule scan</ConfigButton>
                {showScheduleScanForm && <ScheduleScanForm namespaces={namespaces} onClose={() => setShowScheduleScanForm(false)} />}
            </div>
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