import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabMisconfigurationDetails from './TabMisconfigurationDetails';

const MISCONFIGURATION_DETAILS_PATHS = {
    MISCONFIGURATION_DETAILS: "",
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
                    component: () => <TabMisconfigurationDetails data={data} />
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
