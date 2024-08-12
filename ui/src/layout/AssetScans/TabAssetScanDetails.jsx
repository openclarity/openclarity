import React, { useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import { isEmpty } from 'lodash';
import TitleValueDisplay, { TitleValueDisplayRow } from 'components/TitleValueDisplay';
import DoublePaneDisplay from 'components/DoublePaneDisplay';
import Title from 'components/Title';
import ErrorMessageDisplay from 'components/ErrorMessageDisplay';
import { WrappingTextBoxWithEllipsis } from 'components/WrappingTextBoxWithEllipsis';
import { ROUTES, FINDINGS_MAPPING, VULNERABIITY_FINDINGS_ITEM } from 'utils/systemConsts';
import { formatDate, calculateDuration } from 'utils/utils';
import { SCANS_PATHS } from 'layout/Scans';
import StatusIndicator from './StatusIndicator';

import COLORS from 'utils/scss_variables.module.scss';

const BORDER_COLOR = COLORS["color-grey-lighter"];

const STATUS_DISPLAY_ITEMS = [
    {dataKey: "sbom", title: "SBOM"},
    VULNERABIITY_FINDINGS_ITEM,
    ...Object.values(FINDINGS_MAPPING).filter(({value}) => value !== FINDINGS_MAPPING.PACKAGES.value)
]

const StatusDisplay = ({state, errors}) => (
    <>
        <div style={{marginBottom: "20px"}}>
            <StatusIndicator state={state} errors={errors} />
        </div>
        {!isEmpty(errors) &&
            <ErrorMessageDisplay title="An error has occured">
                {errors.map((error, index) => <div key={index}>{error}</div>)}
            </ErrorMessageDisplay>
        }
    </>
)

const TimeDataDisplayRow = ({startTime, endTime}) => (
    <TitleValueDisplayRow>
        <TitleValueDisplay title="Started">{formatDate(startTime)}</TitleValueDisplay>
        <TitleValueDisplay title="Ended">{formatDate(endTime)}</TitleValueDisplay>
        <TitleValueDisplay title="Duration">{calculateDuration(startTime, endTime)}</TitleValueDisplay>
    </TitleValueDisplayRow>
)

const RootFsDisplay = ({ path, size, type }) =>
    <TitleValueDisplayRow>
        <TitleValueDisplay title="Path">{path}</TitleValueDisplay>
        <TitleValueDisplay title="Size">{size}</TitleValueDisplay>
        <TitleValueDisplay title="Type">{type}</TitleValueDisplay>
    </TitleValueDisplayRow>

const CsvDisplay = ({path: input, type}) =>
    <TitleValueDisplayRow>
        <TitleValueDisplay title="Input">
            <WrappingTextBoxWithEllipsis numberOfLines={5}>
                {input}
            </WrappingTextBoxWithEllipsis>
        </TitleValueDisplay>
        <TitleValueDisplay title="Count">{(input || '').split(',').length.toString()}</TitleValueDisplay>
        <TitleValueDisplay title="Type">{type}</TitleValueDisplay>
    </TitleValueDisplayRow>

const StatsDisplay = ({scanTime, ...props}) => {
    const {startTime, endTime} = scanTime || {};

    const statsByType = useMemo(() => {
        switch (props.type) {
            case 'rootfs':
                return <RootFsDisplay {...props} />
            case 'csv':
                return <CsvDisplay {...props} />
            default:
                return <RootFsDisplay {...props} />
        }
    }, [props]);

    return (
        <div style={{border: `1px solid ${BORDER_COLOR}`, borderBottom: "none", padding: "15px"}}>
            {statsByType}
            <TimeDataDisplayRow startTime={startTime} endTime={endTime} />
        </div>
    )
}

const TabAssetScanDetails = ({data}) => {
    const navigate = useNavigate();

    const {scan, asset, status, stats} = data || {};
    const {id: assetId, assetInfo} = asset || {};
    const {instanceID, objectType, location} = assetInfo || {};
    const {id: scanId, startTime, endTime} = scan || {};
    const {state, message} = status || {};

    const ITEM_MARGIN = "46px";

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
                    <TimeDataDisplayRow startTime={startTime} endTime={endTime} />
                </>
            )}
            rightPlaneDisplay={() => (
                <>
                    <Title medium>Asset scan details</Title>
                    <TitleValueDisplay title="Overview" isLargeTitle>
                        <StatusDisplay state={state} errors={message === undefined ? [] : [message]} />
                        <TimeDataDisplayRow startTime={startTime} endTime={endTime} />
                    </TitleValueDisplay>
                    <div style={{borderBottom: `3px solid ${BORDER_COLOR}`, margin: "20px 0"}}></div>
                    {
                        STATUS_DISPLAY_ITEMS.map(({dataKey, title}, index) => {
                            const typeStats = (stats || {})[dataKey] || [];
                            const {status: { state, message }} = (data || {})[dataKey] || {status: {}};

                            return (
                                <div key={dataKey} style={{marginTop: ITEM_MARGIN}}>
                                    <TitleValueDisplay title={!!title ? `${title} scan details` : null} isLargeTitle>
                                        <div style={{marginBottom: ITEM_MARGIN}}><StatusDisplay state={state} errors={message === undefined ? [] : [message]} /></div>
                                        {typeStats?.map((typeStats, index) => <StatsDisplay key={index} {...typeStats} />)}
                                    </TitleValueDisplay>
                                    {index + 1 < STATUS_DISPLAY_ITEMS.length &&
                                        <div style={{borderBottom: `1px solid ${BORDER_COLOR}`}}></div>
                                    }
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
