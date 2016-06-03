import React, { PropTypes, Component } from 'react'
import moment from 'moment'
import { levelToString } from "../lib"

class Log extends Component {
    render() {
        let msg = this.props.entry
        let data = []
        for (let key in msg.Data) {
            if (msg.Data.hasOwnProperty(key)) {
                data.push({key: key, value: msg.Data[key]})
            }
        }

        let level = levelToString(msg.Level)
        let labelClass = "label"

        switch (level) {
        case "INFO":
            labelClass += " label-info"
            break
        case "WARN":
            labelClass += " label-warning"
            break
        case "ERROR":
        case "FATAL":
        case "PANIC":
            labelClass += " label-danger"
        default:
            labelClass += " label-default"
        }

        return (
            <div className="row log-entry">
                <div className="col-sm-1">
                    <div className="log-time">
                        {moment(msg.Time).format("MMM D h:mm:ssa")}
                    </div>
                    <span className={"log-level "+labelClass}>{level}</span>
                </div>
                <div className="col-sm-2 log-message">
                    {msg.Message}
                </div>
                <div className="col-sm-9">
                    <div className="log-data">
                        {data.map(function(d, idx) {
                            return (
                                <div className="log-data-pair" key={idx}>
                                    <span className={"log-data-key label-pill "+labelClass}>{d.key}</span>
                                    <div className="log-data-value">{d.value}</div>
                                </div>
                            )
                        })}
                    </div>
                </div>
            </div>
        )
    }
}

Log.propTypes = {
    entry: PropTypes.shape({
        Data: PropTypes.object,
        Level: PropTypes.number.isRequired,
        Time: PropTypes.string.isRequired,
        Message: PropTypes.string.isRequired
    }).isRequired
}

export default Log
