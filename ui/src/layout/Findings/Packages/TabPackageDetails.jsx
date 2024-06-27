import React from 'react';
import TitleValueDisplay, { TitleValueDisplayRow, ValuesListDisplay } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import { FindingsDetailsCommonFields } from '../utils';
import Title from "../../../components/Title/index.jsx";
import LinksList from "../../../components/LinksList/index.jsx";
import {useLocation} from "react-router-dom";
import {FILTER_TYPES, setFilters, useFilterDispatch} from "../../../context/FiltersProvider.js";
import {ROUTES} from "../../../utils/systemConsts.js";
import VulnerabilitiesDisplay from "../../../components/VulnerabilitiesDisplay/index.jsx";

const TabPackageDetails = ({data}) => {
    const {pathname} = useLocation();
    const filtersDispatch = useFilterDispatch();

    const {id, findingInfo, firstSeen, lastSeen, summary} = data;
    const {totalVulnerabilities} = summary || {};
    const {name, version, language, licenses} = findingInfo;

    const onVulnerabilitiesClick = () => {
        setFilters(filtersDispatch, {
            type: FILTER_TYPES.FINDINGS_VULNERABILITIES,
            filters: {
                filter: `findingInfo/package/name eq '${name}' and findingInfo/package/version eq '${version}'`,
                name: `Package ${id}`,
                suffix: "finding",
                backPath: pathname
            },
            isSystem: true
        });
    }

    return (
        <DoublePaneDisplay
            leftPaneDisplay={() => (
                <>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Package name">{name}</TitleValueDisplay>
                        <TitleValueDisplay title="Version">{version}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Language">{language}</TitleValueDisplay>
                        <TitleValueDisplay title="Licenses"><ValuesListDisplay values={licenses} /></TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <FindingsDetailsCommonFields firstSeen={firstSeen} lastSeen={lastSeen} />
                </>  
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium>Package Vulnerabilities</Title>
                    <LinksList
                        items={[
                            {
                                path: ROUTES.FINDINGS,
                                component: () => <VulnerabilitiesDisplay counters={totalVulnerabilities} />,
                                callback: onVulnerabilitiesClick
                            }
                        ]}
                    />
                </>
            )}
        />
    )
}

export default TabPackageDetails;
