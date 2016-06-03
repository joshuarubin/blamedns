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
            <tr className="log-entry">
                <td>
                    <span className={labelClass}>{level}</span>
                </td>
                <td>
                    <small>{moment(msg.Time).format("MMM D h:mm:ssa")}</small>
                </td>
                <td>
                    <small>{msg.Message}</small>
                </td>
                <td>
                    {data.map(function(d, idx) {
                        return (
                            <small className="log-data" key={idx}>
                                <span className={textClass}>{d.key}</span>
                                <span>{d.value}</span>
                            </small>
                        )
                    })}
                </td>
            </tr>
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
