import { create } from 'context/utils';

const initialState = {
    modalDisplayData: null
};

export const MODAL_DISPLAY_ACTIONS = {
    SET_MODAL_DISPLAY_DATA: "SET_MODAL_DISPLAY_DATA",
    CLOSE_DISPLAY_MODAL: "CLOSE_DISPLAY_MODAL"
}

const reducer = (state, action) => {
    switch (action.type) {
        case MODAL_DISPLAY_ACTIONS.SET_MODAL_DISPLAY_DATA: {
            return {
                modalDisplayData: action.payload
            };
        }
        case MODAL_DISPLAY_ACTIONS.CLOSE_DISPLAY_MODAL: {
            return {
                modalDisplayData: null
            };
        }
        default:
            return state;
    }
}

const [ModalDisplayProvider, useModalDisplayState, useModalDisplayDispatch] = create(reducer, initialState);

export {
    ModalDisplayProvider,
    useModalDisplayState,
    useModalDisplayDispatch
};
