import type { ComponentType, ReactNode } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useGetCustomerDashboardQuery } from '@/store/api/appApi';
import { formatEuro } from '@/lib/format';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Skeleton } from '@/components/ui/skeleton';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  BarChart, Bar, LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer,
  CartesianGrid, PieChart, Pie, Cell, Legend,
} from 'recharts';
import type { PieLabelRenderProps } from 'recharts/types/polar/Pie';
import {
  ArrowLeft, ClipboardList, FileText, TrendingUp, MapPin, Layers, Loader2,
} from 'lucide-react';

const PIE_COLORS = ['#0EA5A6', '#0B1220', '#FFAA45', '#22C55E', '#3B82F6', '#A855F7'];

interface KPICardProps {
  title: string;
  value: ReactNode;
  icon: ComponentType<{ className?: string }>;
  description?: string;
}

const KPICard = ({ title, value, icon: Icon, description }: KPICardProps) => (
  <Card className="shadow-sm">
    <CardContent className="p-4 lg:p-5">
      <div className="flex items-start justify-between">
        <div>
          <p className="text-xs font-medium text-muted-foreground mb-1">{title}</p>
          <p className="text-2xl md:text-3xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>{value}</p>
          {description && <p className="text-xs text-muted-foreground mt-1">{description}</p>}
        </div>
        <div className="p-2 rounded-lg bg-accent">
          <Icon className="h-4 w-4 text-primary" />
        </div>
      </div>
    </CardContent>
  </Card>
);

export default function CustomerDashboardPage() {
  const { id } = useParams<{ id: string }>();
  const query = useGetCustomerDashboardQuery(id as string, { skip: !id });

  if (query.isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-72 rounded-xl" />
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
          {[1, 2, 3, 4].map(i => <Skeleton key={i} className="h-24 rounded-xl" />)}
        </div>
        <Skeleton className="h-64 rounded-xl" />
      </div>
    );
  }

  if (query.error) {
    return (
      <div className="p-8 text-center">
        <p className="text-destructive mb-4">Errore caricamento cruscotto cliente</p>
        <Button asChild variant="outline">
          <Link to="/customers"><ArrowLeft className="h-4 w-4 mr-2" /> Torna ai clienti</Link>
        </Button>
      </div>
    );
  }

  const data = query.data;
  if (!data) return null;
  const { customer = {}, kpi = {}, monthly_trend = [], top_destinazioni = [], per_tipologia = [], per_categoria = [] } = data;

  return (
    <div className="space-y-4 lg:space-y-6" data-testid="customer-dashboard-page">
      {/* Header */}
      <div className="flex items-center justify-between gap-2">
        <div>
          <Button asChild variant="ghost" size="sm" className="mb-1 h-7 text-xs gap-1">
            <Link to="/customers"><ArrowLeft className="h-3.5 w-3.5" /> Clienti</Link>
          </Button>
          <h1 className="text-xl md:text-2xl font-bold tracking-tight" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
            {customer.ragione_sociale}
          </h1>
          <p className="text-xs text-muted-foreground">
            {customer.citta || '—'}{customer.partita_iva ? ` · P.IVA ${customer.partita_iva}` : ''}
          </p>
        </div>
        {query.isFetching && <Loader2 className="h-4 w-4 animate-spin text-muted-foreground" />}
      </div>

      {/* KPI */}
      <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 lg:gap-4">
        <KPICard
          title="Ordini Totali"
          value={kpi.ordini_totali}
          icon={ClipboardList}
          description={`${kpi.ordini_pianificabili || 0} da pianificare`}
        />
        <KPICard
          title="In Viaggio"
          value={kpi.ordini_in_viaggio}
          icon={Layers}
          description={`${kpi.ordini_chiusi || 0} chiusi`}
        />
        <KPICard
          title="Fatturati"
          value={kpi.ordini_fatturati}
          icon={FileText}
          description={`Tariffa media € ${formatEuro(kpi.tariffa_media || 0)}`}
        />
        <KPICard
          title="Fatturato Netto"
          value={`€ ${formatEuro(kpi.fatturato_netto || 0)}`}
          icon={TrendingUp}
          description="Ordini con fattura collegata"
        />
      </div>

      {/* Trend mensile + Top destinazioni */}
      <div className="grid grid-cols-1 xl:grid-cols-[1.6fr_1fr] gap-4">
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              <TrendingUp className="h-4 w-4" /> Andamento ultimi 12 mesi
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-56">
              {monthly_trend.length === 0 ? (
                <div className="flex items-center justify-center h-full text-sm text-muted-foreground">
                  Nessun dato disponibile
                </div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={monthly_trend}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(214 18% 88%)" />
                    <XAxis dataKey="mese" tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                    <YAxis yAxisId="left" tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                    <YAxis yAxisId="right" orientation="right" tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                    <Tooltip
                      contentStyle={{ borderRadius: 8, fontSize: 12, border: '1px solid hsl(214 18% 88%)' }}
                      formatter={(val: number, name: string) => name === 'fatturato' ? [`€ ${formatEuro(val)}`, 'Fatturato'] : [val, 'Ordini']}
                    />
                    <Legend wrapperStyle={{ fontSize: 11 }} />
                    <Line yAxisId="left" type="monotone" dataKey="ordini" stroke="#0EA5A6" strokeWidth={2} />
                    <Line yAxisId="right" type="monotone" dataKey="fatturato" stroke="#FFAA45" strokeWidth={2} />
                  </LineChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base flex items-center gap-2" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              <MapPin className="h-4 w-4" /> Top destinazioni
            </CardTitle>
          </CardHeader>
          <CardContent>
            {top_destinazioni.length === 0 ? (
              <p className="text-sm text-muted-foreground py-8 text-center">Nessun dato</p>
            ) : (
              <ul className="space-y-2 text-sm">
                {top_destinazioni.map((d, i) => (
                  <li key={d.nome} className="flex items-center justify-between gap-2 py-1 border-b last:border-b-0">
                    <span className="flex items-center gap-2 truncate">
                      <Badge variant="outline" className="text-[10px] tabular-nums">{i + 1}</Badge>
                      <span className="truncate">{d.nome}</span>
                    </span>
                    <span className="text-xs text-muted-foreground tabular-nums whitespace-nowrap">
                      {d.ordini} ord · € {formatEuro(d.fatturato || 0)}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </CardContent>
        </Card>
      </div>

      {/* Distribuzione per tipologia + categoria */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              Per tipologia
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-48">
              {per_tipologia.length === 0 ? (
                <div className="flex items-center justify-center h-full text-sm text-muted-foreground">Nessun dato</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <PieChart>
                    <Pie data={per_tipologia} dataKey="ordini" nameKey="tipologia" outerRadius={70} label={(e: PieLabelRenderProps) => `${e.name}: ${e.value}`}>
                      {per_tipologia.map((_, i) => (
                        <Cell key={i} fill={PIE_COLORS[i % PIE_COLORS.length]} />
                      ))}
                    </Pie>
                    <Tooltip contentStyle={{ borderRadius: 8, fontSize: 12 }} />
                  </PieChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>

        <Card className="shadow-sm">
          <CardHeader className="pb-2">
            <CardTitle className="text-base" style={{ fontFamily: "'Space Grotesk', sans-serif" }}>
              Per categoria trasporto
            </CardTitle>
          </CardHeader>
          <CardContent>
            <div className="h-48">
              {per_categoria.length === 0 ? (
                <div className="flex items-center justify-center h-full text-sm text-muted-foreground">Nessun dato</div>
              ) : (
                <ResponsiveContainer width="100%" height="100%">
                  <BarChart data={per_categoria} layout="vertical" margin={{ left: 10, right: 20 }}>
                    <CartesianGrid strokeDasharray="3 3" stroke="hsl(214 18% 88%)" />
                    <XAxis type="number" tick={{ fontSize: 11 }} stroke="hsl(215 16% 38%)" />
                    <YAxis type="category" dataKey="categoria" tick={{ fontSize: 11 }} width={100} stroke="hsl(215 16% 38%)" />
                    <Tooltip contentStyle={{ borderRadius: 8, fontSize: 12 }} />
                    <Bar dataKey="ordini" fill="#0EA5A6" radius={[0, 4, 4, 0]} />
                  </BarChart>
                </ResponsiveContainer>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
