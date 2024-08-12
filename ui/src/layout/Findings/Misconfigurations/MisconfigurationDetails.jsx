import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabMisconfigurationDetails from './TabMisconfigurationDetails';
import AssetsForFindingTable from 'layout/Assets/AssetsForFindingTable';

const MISCONFIGURATION_DETAILS_PATHS = {
    MISCONFIGURATION_DETAILS: "",
    ASSET_LIST: "assets",
}

const DetailsContent = ({data}) => {
    const {pathname} = useLocation();
    
    const {id} = data;
    
    return (
        <TabbedPage
            basePath={`${pathname.substring(0, pathname.indexOf(id))}${id}`}
            items={[
                {
                    id: "general",
                    title: "Misconfiguration details",
                    isIndex: true,
                    path: MISCONFIGURATION_DETAILS_PATHS.MISCONFIGURATION_DETAILS,
                    component: () => <TabMisconfigurationDetails data={data} />
                },
                {
                    id: "assets",
                    title: "Assets",
                    path: MISCONFIGURATION_DETAILS_PATHS.ASSET_LIST,
                    component: () => <AssetsForFindingTable findingId={id} />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const MisconfigurationDetails = () => (
    <FindingsDetailsPage
        backTitle="Misconfigurations"
        getTitleData={({findingInfo}) => ({title: findingInfo.id})}
        detailsContent={DetailsContent}
    />
)

export default MisconfigurationDetails;
