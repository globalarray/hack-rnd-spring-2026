import io
import base64
import numpy as np
from matplotlib.figure import Figure  # Используем Figure напрямую для потокобезопасности
from docxtpl import DocxTemplate, InlineImage
from docx.shared import Mm
from jinja2 import Template


def _create_chart(labels, scores, chart_type='radar'):
    fig = Figure(figsize=(5, 4), dpi=120)
    prof_color = '#10b981'

    if chart_type == 'radar':
        ax = fig.add_subplot(111, polar=True)
        # Подготовка данных для замыкания круга
        stats = np.concatenate((scores, [scores[0]]))
        angles = np.linspace(0, 2 * np.pi, len(labels), endpoint=False).tolist()
        angles += angles[:1]

        ax.fill(angles, stats, color=prof_color, alpha=0.2)
        ax.plot(angles, stats, color=prof_color, linewidth=2, marker='o')

        ax.set_xticks(angles[:-1])
        ax.set_xticklabels(labels)
        ax.set_ylim(0, 5)
    else:
        ax = fig.add_subplot(111)
        ax.bar(labels, scores, color=prof_color, alpha=0.8)
        ax.set_ylim(0, 5)

    # Сохраняем через объект фигуры
    buf = io.BytesIO()
    fig.savefig(buf, format='png', bbox_inches='tight')
    buf.seek(0)
    return buf


def render_any_report(template_bytes, data, file_name):
    # Защита от пустых данных, если labels/scores не дошли
    if not data.get('labels') or not data.get('scores'):
        data['labels'], data['scores'] = ['No Data'], [0]

    if file_name.endswith('.docx'):
        doc = DocxTemplate(io.BytesIO(template_bytes))

        # Генерируем изображения
        radar_buf = _create_chart(data['labels'], data['scores'], 'radar')
        bar_buf = _create_chart(data['labels'], data['scores'], 'bar')

        # Вставляем в Word
        data['radar_chart'] = InlineImage(doc, radar_buf, width=Mm(90))
        data['graph'] = InlineImage(doc, bar_buf, width=Mm(120))

        doc.render(data)
        out = io.BytesIO()
        doc.save(out)
        return out.getvalue()

    else:
        # HTML рендеринг
        # Чтобы не генерировать графики дважды, создаем их один раз
        radar_buf = _create_chart(data['labels'], data['scores'], 'radar')
        bar_buf = _create_chart(data['labels'], data['scores'], 'bar')

        data['radar_chart_base64'] = base64.b64encode(radar_buf.getvalue()).decode()
        data['chart_base64'] = base64.b64encode(bar_buf.getvalue()).decode()

        return Template(template_bytes.decode('utf-8')).render(data).encode('utf-8')