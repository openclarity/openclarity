import React from 'react';
import { useNavigate } from 'react-router-dom';
import { isEmpty } from 'lodash';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import ErrorMessageDisplay from 'components/ErrorMessageDisplay';
import { ROUTES, FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { formatDate, calculateDuration } from 'utils/utils';
import { SCANS_PATHS } from 'layout/Scans';
import StatusIndicator from './StatusIndicator';

import COLORS from 'utils/scss_variables.module.scss';

const STATUS_DISPLAY_ITEMS = [
    {dataKey: "sbom", title: "SBOM"},
    VULNERABIITY_FINDINGS_ITEM,
    ...Object.values(FINDINGS_MAPPING).filter(({value}) => value !== FINDINGS_MAPPING.PACKAGES.value)
]

const StatusDisplay = ({state, errors}) => (
    <>
        <StatusIndicator state={state} errors={errors} />
        {!isEmpty(errors) &&
            <div style={{marginTop: "20px"}}>
                <ErrorMessageDisplay title="An error has occured">
                    {errors.map((error, index) => <div key={index}>{error}</div>)}
                </ErrorMessageDisplay>
            </div>
        }
    </>
)

const TabAssetScanDetails = ({data}) => {
    const navigate = useNavigate();

    const {scan, asset, status} = data || {};
    const {id: assetId, assetInfo} = asset || {};
    const {instanceID, objectType, location} = assetInfo || {};
    const {id: scanId, startTime, endTime} = scan || {};
    const {state, errors} = status?.general || {};

    return (
        <DoublePaneDisplay
            className="asset-scans-details-tab-general"
            leftPaneDisplay={() => (
                <>
                    <Title medium onClick={() => navigate(`${ROUTES.ASSETS}/${assetId}`)}>Asset</Title>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Name">{instanceID}</TitleValueDisplay>
                        <TitleValueDisplay title="Type">{objectType}</TitleValueDisplay>
                        <TitleValueDisplay title="Location">{location}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                    <Title medium onClick={() => navigate(`${ROUTES.SCANS}/${SCANS_PATHS.SCANS}/${scanId}`)}>Scan</Title>
                    <TitleValueDisplayRow>
                        <TitleValueDisplay title="Started">{formatDate(startTime)}</TitleValueDisplay>
                        <TitleValueDisplay title="Ended">{formatDate(endTime)}</TitleValueDisplay>
                        <TitleValueDisplay title="Duration">{calculateDuration(startTime, endTime)}</TitleValueDisplay>
                    </TitleValueDisplayRow>
                </>
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium>Asset scan status</Title>
                    <TitleValueDisplay title="Overview">
                        <StatusDisplay state={state} errors={errors} />
                    </TitleValueDisplay>
                    <div style={{borderBottom: `2px solid ${COLORS["color-grey-lighter"]}`, margin: "20px 0"}}></div>
                    {
                        STATUS_DISPLAY_ITEMS.map(({dataKey, title}) => {
                            const {state, errors} = (status || {})[dataKey] || {};

                            return (
                                <div key={dataKey} style={{marginTop: "46px"}}>
                                    <TitleValueDisplay title={!!title ? `${title} scan status` : null}>
                                        <StatusDisplay state={state} errors={errors} />
                                    </TitleValueDisplay>
                                </div>
                            )
                        })
                    }
                </>
            )}
        />
    )
}

export default TabAssetScanDetails;