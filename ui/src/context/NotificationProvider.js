import { create } from './utils';

const initialState = {
    message: null,
    type: null
};

const NOTIFICATION_ACTIONS = {
    SHOW_NOTIFICATION: "SHOW_NOTIFICATION",
    REMOVE_NOTIFICATION: "REMOVE_NOTIFICATION"
};

const reducer = (state, action) => {
    switch (action.type) {
        case NOTIFICATION_ACTIONS.SHOW_NOTIFICATION: {
            const {type, message} = action.payload;
            
            return {
                ...state,
                message,
                type
            };
        }
        case NOTIFICATION_ACTIONS.REMOVE_NOTIFICATION: {
            return {
                ...state,
                message: null,
                type: null
            }
        }
        default:
            return state;
    }
}

const [NotificationProvider, useNotificationState, useNotificationDispatch] = create(reducer, initialState);

const removeNotification = (dispatch) => dispatch({type: NOTIFICATION_ACTIONS.REMOVE_NOTIFICATION});
const showNotification = (dispatch, {type, message}) => {
    dispatch({type: NOTIFICATION_ACTIONS.SHOW_NOTIFICATION, payload: {type, message}});
    setTimeout(() => removeNotification(dispatch), 6000);
};

export {
    NotificationProvider,
    useNotificationState,
    useNotificationDispatch,
    showNotification,
    removeNotification
};