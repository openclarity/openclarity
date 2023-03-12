import React from 'react';
import { useNavigate } from 'react-router-dom';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import Button from 'components/Button';
const SCANS_COUNT = 5;

const TabScans = () => {
    const navigate = useNavigate();
    
    return (
        <DoublePaneDisplay
            className="configuration-details-tab-scans"
            leftPaneDisplay={() => (
                <>
                    <Title medium>{`Last ${SCANS_COUNT} scans`}</Title>
                    <div className="configuration-details-tab-scans-list">
                        <Button tertiary onClick={() => navigate(0)}>placeholder</Button>
                    </div>
                    <Button onClick={() => navigate(0)}>See all scans (?)</Button>
                </>
            )}
        />
    )
}

export default TabScans;
