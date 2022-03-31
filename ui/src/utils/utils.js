import moment from 'moment';

export const formatDateBy = (date, format) => !!date ? moment(date).format(format): "";
export const formatDate = (date) => formatDateBy(date, "MMM Do, YYYY HH:mm:ss");

export const getItemsString = items => {
    if (!items) {
        return null;
    }

    return (
        items.join(", ")
    )
};

export const toCapitalized = string => string.charAt(0).toUpperCase() + string.slice(1).toLowerCase();

export const BoldText = ({children}) => <span style={{fontWeight: "bold"}}>{children}</span>;