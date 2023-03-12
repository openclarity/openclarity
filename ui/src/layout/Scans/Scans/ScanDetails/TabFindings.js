import React from 'react';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import LinksList from 'components/LinksList';
import { FINDINGS_MAPPING } from '../utils';
import VulnerabilitiesDisplay from '../VulnerabilitiesDisplay';
import FindingsCounterDisplay from './FindingsCounterDisplay';

const TabFindings = ({data}) => {
    const {summary} = data || {};
    
    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <Title medium>Findings</Title>
                    <LinksList
                        items={[
                            {path: 0, component: () => <VulnerabilitiesDisplay counters={summary?.totalVulnerabilities} />},
                            ...Object.keys(FINDINGS_MAPPING).map(findingType => {
                                const {key, title, icon} = FINDINGS_MAPPING[findingType];

                                return {
                                    path: 0,
                                    component: () => <FindingsCounterDisplay key={key} icon={icon} count={summary[key]} title={title} />
                                }
                            })
                        ]}
                    />
                </>
            )}
        />
    )
}

export default TabFindings;