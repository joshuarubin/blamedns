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
        let textClass = ""

        switch (level) {
        case "INFO":
            labelClass += " label-info"
            textClass = "text-info"
            break
        case "WARN":
            labelClass += " label-warning"
            textClass = "text-warning"
            break
        case "ERROR":
        case "FATAL":
        case "PANIC":
            labelClass += " label-danger"
            textClass = "text-danger"
        default:
            labelClass += " label-default"
            textClass = "text-muted"
        }

        return (
            <div className="row log-entry">
                <div className="col-sm-1">
                    <span className={"log-level "+labelClass}>{level}</span>
                </div>
                <div className="col-sm-1">
                    <div className="log-time">
                        <small>{moment(msg.Time).format("MMM D h:mm:ssa")}</small>
                    </div>
                </div>
                <div className="col-sm-2">
                    <small><span className="log-message">{msg.Message}</span></small>
                </div>
                <div className="col-sm-8">
                    <div className="log-data">
                        {data.map(function(d, idx) {
                            return (
                                <div className="log-data-pair" key={idx}>
                                    <div className={"log-data-key "+textClass}><small>{d.key}</small></div>
                                    <div className="log-data-value"><small>{d.value}</small></div>
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
