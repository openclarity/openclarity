import cloneDeep from "lodash/cloneDeep";

const convertDataType = (originalType) => {
    switch (originalType) {
        case "string":
            return "text";
        case "integer":
            return "number";
        case "object":
            return "!struct";
        case "array":
            return "!struct"
        default:
            return originalType;
    }
}

const collectProperties = (assetObject) => {
    console.log('assetObject:', cloneDeep(assetObject));

    const KEY_TO_FIND = "properties";
    const propertyList = {};

    const prepareProperties = (unknownValue, targetObject) => {
        if (Array.isArray(unknownValue)) {
            unknownValue.forEach((subElement) => {
                prepareProperties(subElement, targetObject);
            })
        } else if (typeof unknownValue === 'object') {
            for (const [key, value] of Object.entries(unknownValue)) {
                if (key === KEY_TO_FIND && typeof value === 'object' && !Array.isArray(value)) {
                    Object.assign(targetObject, value)
                } else {
                    prepareProperties(value, targetObject);
                }
            }
        }
    }

    const createSubfields = (source, targetObject) => {
        if (Array.isArray(source)) {
            source.forEach((subElement) => {
                createSubfields(subElement, targetObject);
            })
        } else if (typeof source === 'object') {
            for (const [key, value] of Object.entries(source)) {
                if (key === KEY_TO_FIND && typeof value === 'object' && !Array.isArray(value)) {
                    //value here is always the properties object
                    if (targetObject.hasOwnProperty("subfields")) {
                        Object.assign(targetObject.subfields, value)
                    } else {
                        targetObject.subfields = value;
                    }
                    for (const [subKey, subValue] of Object.entries(value)) {
                        createSubfields(subValue, targetObject.subfields[subKey]);
                    }
                } else {
                    createSubfields(value, targetObject);
                }
            }
        }
    }

    const convertDataForQuery = (propertyList) => {
        if (Array.isArray(propertyList)) {
            propertyList.forEach((subElement) => {
                convertDataForQuery(subElement);
            })
        } else if (typeof propertyList === 'object') {
            for (const [key, value] of Object.entries(propertyList)) {
                if (Object.keys(value).includes("type")) {
                    value.type = convertDataType(value.type);
                    value.label = key;

                    if (value.enum) {
                        value.type = 'select';
                        value.fieldSettings = {
                            listValues: value.enum.map(opt => ({
                                title: opt,
                                value: opt
                            }))
                        }
                    }

                    if (value.type === 'text' && value.format === 'date-time') {
                        value.type = 'datetime';
                    }
                }

                if (Object.keys(value).includes("subfields")) {
                    propertyList[`${key}-object`] = {
                        type: 'group-select',
                        label: `${key} (object)`,
                        fieldSettings: {
                            listValues: [{
                                title: 'null',
                                value: 'null'
                            }]
                        },
                        fieldName: key,
                        defaultValue: "null",
                    }
                }

                convertDataForQuery(value);
            }
        }
    }

    // Main properties collected
    prepareProperties(cloneDeep(assetObject), propertyList);
    // child objects are collected under subfields key 
    Object.values(propertyList).forEach(value => createSubfields(value, value))
    // Data conversion (js --> odata)
    convertDataForQuery(propertyList);

    console.log('propertyList:', propertyList);

    return propertyList;
}

export { collectProperties };
