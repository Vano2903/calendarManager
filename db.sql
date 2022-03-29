CREATE TABLE IF NOT EXISTS `sheets` (
    `sheetID` VARCHAR(50) NOT NULL,
    `emailOwner` VARCHAR(75) NOT NULL,
    PRIMARY KEY (`emailOwner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `calendars` (
    `calendarID` VARCHAR(70) NOT NULL,
    `emailOwner` VARCHAR(75) NOT NULL,
    PRIMARY KEY (`emailOwner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `states` (
    `value` VARCHAR(20) NOT NULL,
    `expiration` TIMESTAMP NOT NULL,
    PRIMARY KEY (`value`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `tokens` (
    `ID` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `email` varchar(75) not null,
    `accID` int NOT NULL,
    `refID` int NOT NULL,
    `googleID` int NOT NULL,
    PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `accessTokens` (
    `ID` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `accToken` char(64) NOT NULL,
    `accExp` DATETIME NOT NULL,
    PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `refreshTokens` (
    `ID` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `refreshToken` char(64) NOT NULL,
    `refreshExp` DATETIME NOT NULL,
    PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;

CREATE TABLE IF NOT EXISTS `googleTokens` (
    `ID` int(11) unsigned NOT NULL AUTO_INCREMENT,
    `googleAccessToken` text NOT NULL,
    `googleExp` DATETIME NOT NULL,
    `googleRefreshToken` text NOT NULL,
    PRIMARY KEY (`ID`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8;
