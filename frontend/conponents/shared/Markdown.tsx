import MarkdownIt from "markdown-it";
import { h } from "preact";

const htmlMarkdownIt = new MarkdownIt({
    linkify: true,
    html: true,
});

const nonHtmlMarkdownIt = new MarkdownIt({
    linkify: true,
    html: false,
});

type Props = {
    unsafe: boolean;
    content: string;
};

export default (props: Props) => {
    return <span
        dangerouslySetInnerHTML={{
            __html: (props.unsafe ? htmlMarkdownIt : nonHtmlMarkdownIt).render(props.content)
        }}
    />;
};
