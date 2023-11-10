type InputProps = {
    name: string;
    required: boolean;
}

type Textfield = InputProps & {
    type: "text";
    title: string;
    defaut?: string;
    placeholder?: string;
}

type TextArea = InputProps & {
    type: "textarea";
    title: string;
    defaut?: string;
    placeholder?: string;
}

type Password = InputProps & {
    type: "password";
    title: string;
    defaut?: string;
    placeholder?: string;
}

type Checkbox = InputProps & {
    type: "checkbox";
    label: string;
    title?: string;
    defaut?: boolean;
}

export type Input = Textfield | Password | Checkbox;
