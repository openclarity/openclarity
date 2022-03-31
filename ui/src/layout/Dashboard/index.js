import React, { useState } from 'react';
import TopBarTitle from 'components/TopBarTitle';
import FixableVulnerabilitiesWidget from './FixableVulnerabilitiesWidget';
import TopVulnerabilitiesWidget from './TopVulnerabilitiesWidget';
import PackagesPieWidget from './PackagesPieWidget';
import VulnerabilitiesTrendWidget from './VulnerabilitiesTrendWidget';
import CountersDisplay from './CountersDisplay';

import './dashboard.scss';

const WidgentsLine = ({children}) => <div className="dashboard-widgets-line">{children}</div>

const Dashboard = () => {
    const [refreshTimestamp, setRefreshTimestamp] = useState(Date());
    const doRefreshTimestamp = () => setRefreshTimestamp(Date());

    return (
        <div className="dashboard-page-wrapper">
            <TopBarTitle title="Dahsboard" onRefresh={doRefreshTimestamp} />
            <WidgentsLine>
                <FixableVulnerabilitiesWidget refreshTimestamp={refreshTimestamp} />
                <PackagesPieWidget
                    title="Packages count per license type"
                    url="dashboard/packagesPerLicense"
                    itemTitleKey="license"
                    filterName="license"
                    refreshTimestamp={refreshTimestamp}
                />
                <PackagesPieWidget
                    title="Packages count per language"
                    url="dashboard/packagesPerLanguage"
                    itemTitleKey="language"
                    filterName="language"
                    refreshTimestamp={refreshTimestamp}
                />
                <CountersDisplay refreshTimestamp={refreshTimestamp} />
            </WidgentsLine>
            <WidgentsLine>
                <TopVulnerabilitiesWidget refreshTimestamp={refreshTimestamp} />
                <VulnerabilitiesTrendWidget refreshTimestamp={refreshTimestamp} />
            </WidgentsLine>
        </div>
    )
}

export default Dashboard;