import React, { useEffect } from 'react';
import { usePrevious, useFetch } from 'hooks';
import { TextField, RadioField, MultiselectField, SelectField, FieldsPair, useFormikContext, CheckboxField,
    FieldLabel, validators } from 'components/Form';

const DISCOVERY_URL = "discovery/scopes";

const idToValueLabel = items => items?.map(({id}) => ({value: id, label: id}));

export const VPCS_EMPTY_VALUE = [{id: "", securityGroups: []}]
export const REGIONS_EMPTY_VALUE = [{id: "", vpcs: VPCS_EMPTY_VALUE}];

export const SCOPE_ITEMS = {
    ALL: {value: "ALL", label: "All"},
    DEFINED: {value: "DEFINED", label: "Define scope"}
}

const SecurityGroupField = ({index, name, placeholder, disabled, regionData}) => {
    const {id: regionId, vpcs=[]} = regionData || {};
    const vpcId = vpcs[index]?.id;
    const prevVpsId = usePrevious(vpcId);

    const [{data, loading, error}, fetchSecurityGroups] = useFetch(DISCOVERY_URL, {loadOnMount: false});

    useEffect(() => {
        if (!!vpcId && prevVpsId !== vpcId) {
            fetchSecurityGroups({queryParams: {
                "$select": "AwsScope.Regions.Vpcs.securityGroups",
                "$filter": `AwsScope.Regions.ID eq '${regionId}' and AwsScope.Regions.Vpcs.ID eq '${vpcId}'`
            }});
        }
    }, [prevVpsId, vpcId, regionId, fetchSecurityGroups]);

    if (error) {
        return null;
    }

    const vpcData = data?.regions[0]?.vpcs[0];

    return (
        <MultiselectField
            name={name}
            placeholder={placeholder}
            disabled={disabled}
            items={!!data ? idToValueLabel(vpcData?.securityGroups) : []}
            loading={loading}
        />
    )
}

const RegionFields = ({index, name, disabled}) => {
    const {values} = useFormikContext();
    const {regions} = values.scope;
    
    const regionId = regions[index]?.id;
    const prevRegionId = usePrevious(regionId);

    const [{data, loading, error}, fetchVpcs] = useFetch(DISCOVERY_URL, {loadOnMount: false});

    useEffect(() => {
        if (!!regionId && prevRegionId !== regionId) {
            fetchVpcs({queryParams: {
                "$select": "AwsScope.Regions.Vpcs",
                "$filter": `AwsScope.Regions.ID eq '${regionId}'`
            }});
        }
    }, [prevRegionId, regionId, fetchVpcs]);

    if (error) {
        return null;
    }
    
    return (
        <FieldsPair
            name={name}
            disabled={disabled}
            firstFieldProps={{
                component: SelectField,
                key: "id",
                placeholder: "Select VPC...",
                items: !!data ? idToValueLabel(data?.regions[0]?.vpcs) : [],
                clearable: true,
                loading: loading
            }}
            secondFieldProps={{
                component: SecurityGroupField,
                key: "securityGroups",
                placeholder: "Select security group...",
                emptyValue: [],
                regionData: regions[index]
            }}
        />
    );
}

const DefinedScopeFields = () => {
    const [{data, loading, error}] = useFetch(DISCOVERY_URL, {queryParams: {"$select": "AwsScope.Regions"}});

    if (error) {
        return null;
    }

    return (
        <FieldsPair
            name="scope.regions"
            firstFieldProps={{
                component: SelectField,
                key: "id",
                placeholder: "Select region...",
                items: idToValueLabel(data?.regions),
                validate: validators.validateRequired,
                loading: loading
            }}
            secondFieldProps={{
                component: RegionFields,
                key: "vpcs",
                emptyValue: VPCS_EMPTY_VALUE
            }}
        />
    )
}

const StepGeneralProperties = () => {
    const {values, setFieldValue} = useFormikContext();
    const {scopeSelect} = values.scope;

    const isDefineScope = scopeSelect === SCOPE_ITEMS.DEFINED.value;
    const prevIsDefineScope = usePrevious(isDefineScope);
    
    useEffect(() => {
        if (prevIsDefineScope && !isDefineScope) {
            setFieldValue("scope.regions", REGIONS_EMPTY_VALUE);
        }
    }, [prevIsDefineScope, isDefineScope, setFieldValue]);
    
    return (
        <div className="scan-config-general-step">
            <TextField
                name="name"
                label="Scan config name*"
                placeholder="Type a name..."
                validate={validators.validateRequired}
            />
            <RadioField
                name="scope.scopeSelect"
                label="Scope"
                items={Object.values(SCOPE_ITEMS)}
            />
            {isDefineScope && <DefinedScopeFields />}
            <div className="scan-config-instances">
                <FieldLabel>Instances</FieldLabel>
                <CheckboxField
                    name="scope.shouldScanStoppedInstances"
                    title="Scan non-running instances"
                />
                <MultiselectField
                    name="scope.instanceTagSelector"
                    placeholder="Select instances to include (key=value)..."
                    connector="and"
                    prefixLabel="Include"
                    creatable
                    validate={validators.keyValueListValidator}
                />
                <MultiselectField
                    name="scope.instanceTagExclusion"
                    placeholder="Select instances to exclude (key=value)..."
                    connector="and"
                    prefixLabel="Exclude"
                    creatable
                    validate={validators.keyValueListValidator}
                />
            </div>
        </div>
    )
}

export default StepGeneralProperties;