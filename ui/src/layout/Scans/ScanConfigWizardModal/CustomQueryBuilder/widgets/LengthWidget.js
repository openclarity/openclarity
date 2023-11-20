import React, { useCallback } from "react";
//import { AntdWidgets } from '@react-awesome-query-builder/antd';
import { InputNumber, Col } from "antd";

// const {
//     FieldSelect,
//     TextWidget,
//     NumberWidget,
// } = AntdWidgets;

const LengthWidget = (props) => {

    const {
        setValue,
        config,
        //placeholders,
        customProps = {},
        value,
        min,
        max,
        step,
        // marks,
        // textSeparators,
        readonly,
    } = props;

    const { renderSize } = config.settings;
    const { width, ...rest } = customProps || {};
    const customInputProps = rest.input || {};

    //const operator = value && Array.isArray(value) && value[0];
    const numValue = value && Array.isArray(value) && value[1];

    /*
    const handleChangeOperator = useCallback((operatorvalue) => {
        if (value && Array.isArray(value)) {
            setValue(prevVal => ([operatorvalue, prevVal[1]]));
        }
    }, [setValue, value]);
    */
    const handleChangeNumberValue = useCallback((numValue) => {
        if (value && Array.isArray(value)) {
            setValue(prevVal => ([prevVal[0], numValue]));
        }
    }, [setValue, value]);

    return (
        <Col style={{ display: "inline-flex", flexWrap: "wrap" }}>
            <Col style={{ float: "left", marginRight: "5px" }}>
                {/* Input for operators */}
            </Col>
            <Col style={{ float: "left", marginRight: "5px" }}>
                <InputNumber
                    disabled={readonly}
                    size={renderSize}
                    key="lengthVal"
                    value={numValue}
                    min={min}
                    max={max}
                    step={step}
                    //placeholder={placeholders[1]}
                    onChange={handleChangeNumberValue}
                    {...customInputProps}
                />
            </Col>
            <Col style={{ clear: "both" }} />
        </Col>
    );
};

export default LengthWidget;