import React from 'react';
import classnames from 'classnames';
import { Cron } from 'react-js-cron';
import { useField } from 'formik';
import Tag from 'components/Tag';
import { cronExpressionToHuman } from 'utils/utils';

import './cron-field.scss';

const CronTitle = ({children}) => <div className="cron-field-title">{children}</div>;

const CronField = (props) => {
    const {className, quickOptions=[]} = props;
    // eslint-disable-next-line no-unused-vars
    const [field, meta, helpers] = useField(props);
    const {value} = field; 
    const {setValue, setTouched} = helpers;
    
    return (
        <div className={classnames("form-field-wrapper", "cron-field-wrapper", {[className]: className})} onBlur={() => setTouched(true, true)}>
            {!!quickOptions &&
                <div className="cron-field-quick-options-wrapper">
                    <CronTitle>Quick options</CronTitle>
                    <div className="cron-field-quick-options">
                        {quickOptions.map(({value, label}, index) => (
                            <Tag key={index} onClick={() => setValue(value)}>{label}</Tag>
                        ))}
                    </div>
                    <div className="cron-field-quick-separator">
                        <div className="cron-field-quick-separator-line"></div>
                        <div className="cron-field-quick-separator-text">or</div>
                        <div className="cron-field-quick-separator-line"></div>
                    </div>
                </div>
            }
            <CronTitle>Every</CronTitle>
            <Cron
                value={value}
                setValue={setValue}
                humanizeLabels
                clearButton={false}
                clockFormat="24-hour-clock"
                className="cron-field-select"
            />
            {!!value &&
                <div className="cron-field-output-wrapper">
                    <div>{cronExpressionToHuman(value)}</div>
                    <div className="cron-field-output-expression">{`Cron expression: ${value}`}</div>
                </div>
            }
        </div>
    )
}

export default CronField;