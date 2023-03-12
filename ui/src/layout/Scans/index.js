import React from 'react';
import { useLocation } from 'react-router-dom';
import { isNull } from 'lodash';
import { useMountMultiFetch } from 'hooks';
import TabbedPage from 'components/TabbedPage';
import Loader from 'components/Loader';
import EmptyDisplay from 'components/EmptyDisplay';
import { APIS } from 'utils/systemConsts';
import ScanConfigWizardModal from './ScanConfigWizardModal';
import { ModalDisplayProvider, useModalDisplayState, useModalDisplayDispatch,
    MODAL_DISPLAY_ACTIONS } from './ScanConfigWizardModal/ModalDisplayProvider';
import Scans, { SCAN_SCANS_PATH } from './Scans';
import Configurations, { SCAN_CONFIGS_PATH } from './Configurations';

const ScansTabbedPage = () => {
    const {pathname} = useLocation();

    const {modalDisplayData} = useModalDisplayState();
    const modalDisplayDispatch = useModalDisplayDispatch();
    const closeDisplayModal = () => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.CLOSE_DISPLAY_MODAL});

    const [{data, error, loading}, fetchData] = useMountMultiFetch([
        {key: "scans", url: APIS.SCANS},
        {key: "scanConfigs", url: APIS.SCAN_CONFIGS}
    ]);

    if (loading) {
        return <Loader />;
    }

    if (error) {
        return null;
    }
    
    const {scans, scanConfigs} = data;

    return (
        <>
            {(!scans?.total && !scanConfigs?.total) ?
                <EmptyDisplay
                    message={(
                        <>
                            <div>No scans detected.</div>
                            <div>Create your first scan configuration to see your VM's issues.</div>
                        </>
                    )}
                    title="New scan configuration"
                    onClick={() => modalDisplayDispatch({type: MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA, payload: {}})}
                /> :
                <TabbedPage
                    redirectTo={`${pathname}/${SCAN_SCANS_PATH}`}
                    items={[
                        {
                            id: "scans",
                            title: "Scans",
                            path: SCAN_SCANS_PATH,
                            component: Scans
                        },
                        {
                            id: "configs",
                            title: "Configurations",
                            path: SCAN_CONFIGS_PATH,
                            component: Configurations
                        }
                    ]}
                    withStickyTabs
                />
            }
            {!isNull(modalDisplayData) && 
                <ScanConfigWizardModal
                    initialData={modalDisplayData}
                    onClose={closeDisplayModal}
                    onSubmitSuccess={() => {
                        closeDisplayModal();
                        fetchData();
                    }}
                />
            }
        </>
    )
}

const ScansWrapper = () => (
    <ModalDisplayProvider>
        <ScansTabbedPage />
    </ModalDisplayProvider>
)

export default ScansWrapper;