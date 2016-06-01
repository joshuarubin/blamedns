export const levelToString = (level) => {
    switch (level) {
    case 0:
        return "PANIC"
    case 1:
        return "FATAL"
    case 2:
        return "ERROR"
    case 3:
        return "WARN"
    case 4:
        return "INFO"
    case 5:
        return "DEBUG"
    }
}

export const stringToLevel = (str) => {
    switch (str.toUpperCase()) {
    case "PANIC":
        return 0;
    case "FATAL":
        return 1;
    case "ERROR":
        return 2;
    case "WARN":
        return 3;
    case "INFO":
        return 4;
    case "DEBUG":
        return 5;
    }
}
