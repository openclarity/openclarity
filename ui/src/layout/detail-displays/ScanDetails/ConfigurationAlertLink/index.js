import React from 'react';
import { useNavigate } from 'react-router-dom';
import { isEqual } from 'lodash';
import Title from 'components/Title';
import { ICON_NAMES } from 'components/Icon';
import IconWithTooltip from 'components/IconWithTooltip';
import { SCANS_PATHS } from 'layout/Scans';
import { ROUTES } from 'utils/systemConsts';

import './configuration-alert-link.scss';

const CONFIGURATION_ALERT_TEXT = (
    <span>
        Configuration has been modified since<br />
        the scan has performed and it might not<br />
        match the scan's configuration<br />
    </span>
)

const ConfigurationAlertLink = ({scanConfigData, updatedConfigData}) => {
    const navigate = useNavigate();

    const {id, ...dataToCompare} = updatedConfigData || {};

    return (
        <div className="configuration-alert-link">
            <Title medium removeMargin onClick={() => navigate(`${ROUTES.SCANS}/${SCANS_PATHS.CONFIGURATIONS}/${id}`)}>Configuration</Title>
            {!isEqual(dataToCompare, scanConfigData) &&
                <IconWithTooltip
                    tooltipId="configuration-alert-tooltip"
                    tooltipText={CONFIGURATION_ALERT_TEXT}
                    name={ICON_NAMES.WARNING} 
                />
            }
        </div>
    )
}

export default ConfigurationAlertLink;