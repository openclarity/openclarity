import React from 'react';
import { useLocation } from 'react-router-dom';
import TabbedPage from 'components/TabbedPage';
import FindingsDetailsPage from '../FindingsDetailsPage';
import TabSecretDetails from './TabSecretDetails';

const SECRET_DETAILS_PATHS = {
    SECRET_DETAILS: "",
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
                    title: "Secret details",
                    isIndex: true,
                    component: () => <TabSecretDetails data={data} />
                }
            ]}
            withInnerPadding={false}
        />
    )
}

const SecretDetails = () => (
    <FindingsDetailsPage
        backTitle="Secrets"
        getTitleData={({findingInfo}) => ({title: findingInfo.fingerprint})}
        detailsContent={DetailsContent}
    />
)

export default SecretDetails;
