import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import remarkMath from 'remark-math';
import rehypeKatex from 'rehype-katex';
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter';
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism';
import 'katex/dist/katex.min.css';
import './MarkdownRenderer.scss';

interface MarkdownRendererProps {
  content: string;
}

export function MarkdownRenderer({ content }: MarkdownRendererProps) {
  return (
    <div className="markdown">
      <ReactMarkdown
        remarkPlugins={[remarkGfm, remarkMath]}
        rehypePlugins={[rehypeKatex]}
        components={{
          code({ className, children, ...props }) {
            const match = /language-(\w+)/.exec(className || '');
            const codeString = String(children).replace(/\n$/, '');
            const inline = !match && !codeString.includes('\n');

            if (inline) {
              return (
                <code className="markdown__code-inline" {...props}>
                  {children}
                </code>
              );
            }

            return (
              <div className="markdown__code-block">
                <div className="markdown__code-header">
                  <span>{match?.[1] || 'code'}</span>
                </div>
                <SyntaxHighlighter
                  language={match?.[1] || 'text'}
                  style={oneDark}
                  customStyle={{
                    margin: 0,
                    borderRadius: 0,
                    fontSize: '0.8rem',
                    padding: '0.75rem 1rem',
                  }}
                >
                  {codeString}
                </SyntaxHighlighter>
              </div>
            );
          },
          table({ children }) {
            return (
              <div className="markdown__table-wrapper">
                <table className="markdown__table">{children}</table>
              </div>
            );
          },
          th({ children }) {
            return <th className="markdown__th">{children}</th>;
          },
          td({ children }) {
            return <td className="markdown__td">{children}</td>;
          },
          hr() {
            return null;
          },
          a({ href, children }) {
            return (
              <a
                href={href}
                target="_blank"
                rel="noopener noreferrer"
                className="markdown__link"
              >
                {children}
              </a>
            );
          },
        }}
      >
        {content}
      </ReactMarkdown>
    </div>
  );
}
