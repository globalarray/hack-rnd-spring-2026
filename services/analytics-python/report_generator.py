import io
import numpy as np
import matplotlib.pyplot as plt
from docxtpl import DocxTemplate, InlineImage
from docx.shared import Mm


def _create_chart(labels, scores, chart_type='radar'):
    plt.figure(figsize=(5, 4))
    prof_color = '#10b981'  # Зеленый цвет бренда

    if chart_type == 'radar':
        ax = plt.subplot(111, polar=True)
        stats = np.concatenate((scores, [scores[0]]))
        angles = np.linspace(0, 2 * np.pi, len(labels), endpoint=False).tolist()
        angles += angles[:1]
        ax.fill(angles, stats, color=prof_color, alpha=0.2)
        ax.plot(angles, stats, color=prof_color, linewidth=2, marker='o')
        ax.set_xticks(angles[:-1])
        ax.set_xticklabels(labels)
        ax.set_ylim(0, 5)
    else:
        # Для психологического репорта (тег {{ graph }})
        plt.bar(labels, scores, color=prof_color, alpha=0.8)
        plt.ylim(0, 5)

    buf = io.BytesIO()
    plt.savefig(buf, format='png', bbox_inches='tight', dpi=120)
    plt.close()
    buf.seek(0)
    return buf


def render_any_report(template_bytes, data, file_name):
    if file_name.endswith('.docx'):
        doc = DocxTemplate(io.BytesIO(template_bytes))

        # Репорту пользователя нужен radar_chart
        data['radar_chart'] = InlineImage(doc, _create_chart(data['labels'], data['scores'], 'radar'), width=Mm(90))
        # Репорту психолога нужен graph
        data['graph'] = InlineImage(doc, _create_chart(data['labels'], data['scores'], 'bar'), width=Mm(120))

        doc.render(data)
        out = io.BytesIO()
        doc.save(out)
        return out.getvalue()
    else:
        # HTML оставляем без изменений
        import base64
        from jinja2 import Template
        data['radar_chart_base64'] = base64.b64encode(
            _create_chart(data['labels'], data['scores'], 'radar').getvalue()).decode()
        data['chart_base64'] = base64.b64encode(
            _create_chart(data['labels'], data['scores'], 'bar').getvalue()).decode()
        return Template(template_bytes.decode('utf-8')).render(data).encode('utf-8')
