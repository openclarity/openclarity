import React from 'react';
import { useFetch } from 'hooks';
import Loader from 'components/Loader';
import { APIS } from 'utils/systemConsts';
import { CounterDisplay, ScanCounterDisplay } from './CounterDisplay';
import FindingsTrendsWidget from './FindingsTrendsWidget';
import RiskiestRegionsWidget from './RiskiestRegionsWidget';
import RiskiestAssetsWidget from './RiskiestAssetsWidget';
import FindingsImpactWidget from './FindingsImpactWidget';
import EmptyScansDisplay from './EmptyScansDisplay';

import COLORS from 'utils/scss_variables.module.scss';

import './dashboard.scss';

const COUNTERS_CONFIG = [
    {url: APIS.ASSETS, title: "Assets", background: COLORS["color-gradient-blue"]},
    {url: APIS.FINDINGS, title: "Findings", background: COLORS["color-gradient-yellow"]}
];

const Dashboard = () => {
    const [{data, error, loading}] = useFetch(APIS.SCANS);

    if (loading) {
        return <Loader />;
    }

    if (error) {
        return null;
    }

    if (data.length === 0) {
        return <EmptyScansDisplay />;
    }

    return (
        <div className="dashboard-page-wrapper">
            <ScanCounterDisplay />
            {
                COUNTERS_CONFIG.map(({url, title, background}, index) => (
                    <CounterDisplay key={index} url={url} title={title} background={background} />
                ))
            }
            <RiskiestRegionsWidget className="riskiest-regions" />
            <FindingsTrendsWidget className="findings-trend" />
            <RiskiestAssetsWidget className="riskiest-assets" />
            <FindingsImpactWidget className="findings-impact" />
        </div>
    )
}

export default Dashboard;
