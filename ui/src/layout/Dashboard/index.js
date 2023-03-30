import React from 'react';
import { APIS } from 'utils/systemConsts';
import CounterDisplay from './CounterDisplay';
import FindingsTrendsWidget from './FindingsTrendsWidget';
import RiskiestRegionsWidget from './RiskiestRegionsWidget';
import RiskiestAssetsWidget from './RiskiestAssetsWidget';
import FindingsImpactWidget from './FindingsImpactWidget';

import COLORS from 'utils/scss_variables.module.scss';

import './dashboard.scss';

const COUNTERS_CONFIG = [
    {url: APIS.SCANS, title: "Completed scans", background: COLORS["color-gradient-green"]},
    {url: APIS.ASSETS, title: "Scanned assets", background: COLORS["color-gradient-blue"]},
    {url: APIS.FINDINGS, title: "Risky findings", background: COLORS["color-gradient-yellow"]}
];

const Dashboard = () => {
    return (
        <div className="dashboard-page-wrapper">
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