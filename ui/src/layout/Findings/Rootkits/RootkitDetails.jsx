import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabRootkitDetails from './TabRootkitDetails';
import AssetsForFindingTable from 'layout/Assets/AssetsForFindingTable';

const ROOTKIT_DETAILS_PATHS = {
    ROOTKIT_DETAILS: "",
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
                    title: "Rootkit details",
                    isIndex: true,
                    path: ROOTKIT_DETAILS_PATHS.ROOTKIT_DETAILS,
                    component: () => <TabRootkitDetails data={data} />
                },
                {
                    id: "assets",
                    title: "Assets",
                    path: ROOTKIT_DETAILS_PATHS.ASSET_LIST,
                    component: () => <AssetsForFindingTable findingId={id} />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const RootkitDetails = () => (
    <FindingsDetailsPage
        backTitle="Rootkits"
        getTitleData={({findingInfo}) => ({title: findingInfo.rootkitName})}
        detailsContent={DetailsContent}
    />
)

export default RootkitDetails;
