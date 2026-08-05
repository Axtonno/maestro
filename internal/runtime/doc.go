// Package runtime contiene l'implementazione interna del Runtime Core di Maestro.
//
// I tipi del package nascondono la propria rappresentazione e consentono
// modifiche soltanto attraverso operazioni controllate che esprimono
// un'intenzione di dominio.
//
// Ogni tipo protegge gli invarianti del proprio livello di responsabilità:
//
//   - node mantiene gli invarianti locali di un nodo;
//   - graph mantiene la coerenza tra nodi e relazioni;
//   - registry mantiene la coerenza dei componenti registrati;
//   - resolver costruisce le dipendenze dichiarate;
//   - validator verifica la validità senza modificare il grafo;
//   - builder coordina la costruzione senza esporre risultati parziali;
//   - runtime mantiene la coerenza tra Registry e grafo corrente.
//
// Le strutture mutabili interne non devono essere esposte direttamente.
// Le operazioni che coinvolgono più oggetti devono essere coordinate
// dall'aggregato che ne possiede gli invarianti.
package runtime
