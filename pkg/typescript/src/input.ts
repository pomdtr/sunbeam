type InputProps = {
    title: string;
}

type TextInput = {
    type: "text";
    default?: string;
} & InputProps;

type TextArea = {
    type: "textarea";
    default?: string;
} & InputProps;

type Checkbox = {
    type: "checkbox";
    default?: boolean;
    label?: string;
} & InputProps;

type Select = {
    type: "select";
    default?: string | number | boolean;
    options: Option[];
} & InputProps;

type Option = {
    title: string;
    value: string | number | boolean;
}

export type Input = TextInput | TextArea | Checkbox | Select;
