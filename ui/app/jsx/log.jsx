import React from 'react';
import moment from 'moment';

const levelToString = (level) => {
    switch (level) {
    case 0:
        return "PANIC";
    case 1:
        return "FATAL";
    case 2:
        return "ERROR";
    case 3:
        return "WARN";
    case 4:
        return "INFO";
    case 5:
        return "DEBUG";
    }
}

export default React.createClass({
    propTypes: {
        message: React.PropTypes.object.isRequired,
    },

    render() {
        var msg = this.props.message;
        var data = [];
        for (var key in msg.Data) {
            if (msg.Data.hasOwnProperty(key)) {
                data.push({key: key, value: msg.Data[key]});
            }
        }

        return (
            <div className={"row log-entry log-level-"+msg.Level}>
                <div className="two columns log-time">
                    {moment(msg.Time).format("MMM D h:mm:ssa")}
                </div>
                <div className={"one column log-level log-level-"+msg.Level}>
                    {levelToString(msg.Level)}
                </div>
                <div className="two columns log-message">
                    {msg.Message}
                </div>
                <div className="seven columns log-data">
                    {data.map(function(d, idx) {
                        return (
                            <span key={idx}>
                                <span className="log-data-key">{d.key}</span>
                                <span className="log-data-value">{d.value}</span>
                            </span>
                        );
                    })}
                </div>
            </div>
        );
    }
});
