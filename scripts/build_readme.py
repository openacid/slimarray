#!/usr/bin/env python
# coding: utf-8


import jinja2

def render_j2(tmpl_path, tmpl_vars, output_path):
    template_loader = jinja2.FileSystemLoader(searchpath='./')
    template_env = jinja2.Environment(loader=template_loader,
                                      undefined=jinja2.StrictUndefined)
    template = template_env.get_template(tmpl_path)

    txt = template.render(tmpl_vars)

    with open(output_path, 'w') as f:
        f.write(txt)

if __name__ == "__main__":
    render_j2('docs/README.md.j2', {}, 'README.md')
