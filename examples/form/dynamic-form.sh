#!/bin/bash

sunbeam query --null-input '
{
    type: "form",
    title: "Dynamic Form",
    form: {
        inputs: [
            { name: "name", type: "textfield", title: "textfield" }
        ],
        target: {
            command: "hello"
        }

    }
}
'
