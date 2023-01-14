import { LitElement, html, css } from 'lit';
import { property, customElement } from 'lit/decorators.js';

@customElement('cel-playground')
class CelPlayground extends LitElement {
  @property({ type: String }) expression = '1 < 2';
  @property({ type: String }) result = '';
  @property({ type: String }) error = '';

  render() {
    return html`
      <main>
        <h1>CEL Playground</h1>

        <input @change="${this.updateInput}"/> <button @click="${this.handleEvaluate}">Evaluate</button> 

        ${this.error != ""?
          html`
          <p>
            <pre>${this.error}</pre>
          </p>
          `:
          html`
          <p>
            Result: ${this.result}
          </p>
          `
        }
      </main>
    `;
  }

  updateInput(e: Event) {
    this.expression = (e.target as HTMLInputElement).value;
  }

  async handleEvaluate() {
    this.requestEval(this.expression)
    .then((response) => {
      if (!response.ok) {
         response.text().then(body => {
          this.error = body;
          this.result = "";
        })
      }
      response.json().then((result) => {
        this.error = "";
        this.result = result["result"]
      })
    })
    .catch(error => {
      this.error = error.toString()
      this.result = "";
    })
  }

  async requestEval(expression: String): Promise<Response> {
    const response = await fetch("http://localhost:8080/eval", {
      method: "POST",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({"expression": expression})
    })
    return response
  }
}