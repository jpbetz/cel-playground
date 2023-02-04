import { LitElement, html, css } from 'lit';
import { property, customElement } from 'lit/decorators.js';
import 'prismjs';
import 'lit-code';
import {load} from 'js-yaml';


@customElement('cel-playground')
class CelPlayground extends LitElement {
  @property({ type: String }) expression = '';
  @property({ type: String }) object = '';
  @property({ type: String }) result = '';
  @property({ type: String }) error = '';

  static styles = css`
    button {
      
    }
  `

  constructor() {
    super();
    this.expression = 'object.name == "test"';
    this.object = 'name: test';
  }

  render() {
    return html`
      <main>
        <h1>CEL Playground</h1>

        Object:
        <lit-code
          id="object"
          language='yaml'
          linenumbers
          @update="${this.updateObject}"
        >
        </lit-code>

        CEL:
        <lit-code
          id="expression"
          language='yaml'
          linenumbers
          @update="${this.updateExpression}"
        >
        </lit-code>
        
        <button @click="${this.handleEvaluate}">Evaluate</button> 

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

  updateExpression(litCode: any) {
    this.expression = litCode.detail;
  }
  updateObject(litCode: any) {
    this.object = litCode.detail;
  }

  async handleEvaluate() {
    this.requestEval(this.expression, this.object)
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

  async requestEval(expression: String, object: String): Promise<Response> {
    const response = await fetch("http://localhost:8080/eval", {
      method: "POST",
      headers: {
        "Accept": "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        "expression": expression,
        "variables": {
          "object": load(object.toString()),
        },
    })
    })
    return response
  }
}