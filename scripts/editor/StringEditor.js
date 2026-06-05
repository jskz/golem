/*
 * Copyright (c) 2021 James Skarzinskas.
 * All rights reserved.
 * See LICENSE.txt in project root for license information.
 * Authors:
 *     James Skarzinskas <james@jskarzin.org>
 */
Golem.StringEditor = function (client, string, callback) {
    const splitLines = value => String(value).split(/\r?\n/g);
    const joinLines = lines => lines.join("\r\n");

    let _string = joinLines(splitLines(string));
    let cursor = splitLines(_string).length;

    const commit = () => (string = String(_string));
    const ch = client.character;
    const showEditorBuffer = () => {
        const lines = splitLines(_string);

        if(lines) {
            for(let i = 0; i < lines.length; i++) {
                ch.send((i + 1) + ". " + lines[i] + "\r\n");
            }
        }
    };

    ch.send(
        "\r\n{Y! PRINT | @ DISCARD | $ WRITE | :# GOTO | :d# DELETE LINE | :i# INSERT LINE\r\n" +
        "-[ STRING EDITOR BUFFER ]---------------------------------------------{x\r\n");
    showEditorBuffer();
    ch.send("{Y----------------------------------------------------------------------{x\r\n");

    client.connectionHandler = function (input) {
        try {
            if(input.length > Golem.StringEditor.MaxLineLength) {
                ch.send("{RInput too long, ignoring.{x\r\n");
                return;
            }

            if(!['!', '@', '$', '?'].includes(input) && input[0] !== ':') {
                const lines = splitLines(_string);

                if(lines && lines.length + 1 > Golem.StringEditor.MaxAllowedLines) {
                    ch.send("{RString buffer is full, delete lines to make some room!{x\r\n");
                    return;
                }

                _string = lines
                    ? joinLines(lines.filter((_, index) => index < cursor)
                        .concat([input])
                        .concat(lines.filter((_, index) => index >= cursor)))
                    : input;
                cursor++;
            }

            if(input === '$') {
                commit();
                client.connectionHandler = null;

                callback(client, _string);
            } else if(input === '@') {
                client.connectionHandler = null;

                ch.send("{YExiting without writing changes.{x\r\n");
                callback(client, string);
            } else if(input === '!') {
                ch.send("{YContents of the string editor buffer:{x\r\n");

                showEditorBuffer();
            } else if(Golem.StringEditor.LineNavigationRegex.test(input)) {
                const [_, lineMatch] = Golem.StringEditor.LineNavigationRegex.exec(input);
                const lineNumber = parseInt(lineMatch, 10);

                const lines = splitLines(_string);

                if(lineNumber > lines.length) {
                    ch.send("{R! Cannot move out of bounds (" + lines.length + " lines in buffer){x\r\n");
                    return;
                }

                ch.send("{WOk.  Cursor at " + lineNumber + ". Next line will write to line " + (lineNumber + 1) + "{x\r\n");
                cursor = lineNumber;
            } else if(Golem.StringEditor.DeleteLineRegex.test(input)) {
                const [_, lineMatch] = Golem.StringEditor.DeleteLineRegex.exec(input),
                    lines = splitLines(_string);
                const lineNumber = parseInt(lineMatch, 10);

                if(lineNumber < 1 || lineNumber > lines.length) {
                    ch.send("{R! Cannot delete out of bounds (" + lines.length + " lines in buffer){x\r\n");
                    return;
                }

                if(lines) {
                    const lineIndex = lineNumber - 1;

                    _string = joinLines(lines.filter((_, index) => index !== lineIndex));

                    if(cursor > lineIndex) {
                        cursor--;
                    }

                    cursor = Math.min(cursor, splitLines(_string).length);
                }

                ch.send("{WOk.  Deleted line " + lineNumber + "{x\r\n");
            } else if(Golem.StringEditor.InsertLineRegex.test(input)) {
                const [_, lineMatch] = Golem.StringEditor.InsertLineRegex.exec(input),
                    lines = splitLines(_string);
                const lineNumber = parseInt(lineMatch, 10);

                if(lines) {
                    if(lineNumber < 1 || lineNumber > lines.length + 1) {
                        ch.send("{R! Cannot insert out of bounds (" + lines.length + " lines in buffer){x\r\n");
                        return;
                    }

                    if(lines.length + 1 > Golem.StringEditor.MaxAllowedLines) return;

                    const lineIndex = lineNumber - 1;

                    _string = joinLines(lines
                        .filter((_, index) => index < lineIndex)
                        .concat([""])
                        .concat(lines.filter((_, index) => index >= lineIndex)));

                    if(cursor >= lineIndex) {
                        cursor++;
                    }
                } else {
                    _string = "";
                }

                ch.send("{WOk.  Inserted empty line at line " + lineNumber + "{x\r\n");
            }
        } catch(err) {
            ch.send(err);
        }
    };
};

Golem.StringEditor.MaxAllowedLines = 256;
Golem.StringEditor.MaxLineLength = 256;
Golem.StringEditor.LineNavigationRegex = new RegExp(/^\:(\d+)$/);
Golem.StringEditor.InsertLineRegex = new RegExp(/^\:i(\d+)$/);
Golem.StringEditor.DeleteLineRegex = new RegExp(/^\:d(\d+)$/);
