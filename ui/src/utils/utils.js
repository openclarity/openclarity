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

export const BoldText = ({children, style={}}) => <span style={{fontWeight: "bold", ...style}}>{children}</span>;

export const calculateDuration = (startTime, endTime) => {
    const startMoment = moment(startTime);
    const endMoment = moment(endTime);

    const range = ["days", "hours", "minutes", "seconds"].map(item => ({diff: endMoment.diff(startMoment, item), label: item}))
        .find(({diff}) => diff > 1);

    return !!range ? `${range.diff} ${range.label}` : null;
}